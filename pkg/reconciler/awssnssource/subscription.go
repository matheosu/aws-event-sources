/*
Copyright (c) 2020 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package awssnssource

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/triggermesh/aws-event-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/event"
	"github.com/triggermesh/aws-event-sources/pkg/reconciler/common/skip"
)

// ensureSubscribed ensures the source's HTTP(S) endpoint is subscribed to the
// SNS topic.
func (r *Reconciler) ensureSubscribed(ctx context.Context) error {
	if skip.Skip(ctx) {
		return nil
	}

	src := v1alpha1.SourceFromContext(ctx)
	status := &src.(*v1alpha1.AWSSNSSource).Status

	adapter, err := r.base.FindAdapter(src)
	switch {
	case isNotFound(err):
		return nil
	case err != nil:
		return fmt.Errorf("finding receive adapter: %w", err)
	}

	url := adapter.Status.URL

	// skip this cycle if the adapter URL wasn't yet determined
	if !adapter.IsReady() || url == nil {
		status.MarkNotSubscribed(v1alpha1.AWSSNSReasonNoURL,
			"The receive adapter did not report its public URL yet")
		return nil
	}

	spec := src.(apis.HasSpec).GetUntypedSpec().(v1alpha1.AWSSNSSourceSpec)

	snsClient, err := newSNSClient(r.secretsCli(src.GetNamespace()), spec.ARN.Region, &spec.Credentials)
	if err != nil {
		status.MarkNotSubscribed(v1alpha1.AWSSNSReasonNoClient, "Cannot obtain SNS client")
		return fmt.Errorf("%w", reconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedSubscribe,
			"Error creating SNS client: %s", err))
	}

	resp, err := snsClient.SubscribeWithContext(ctx, &sns.SubscribeInput{
		Endpoint:              aws.String(url.String()),
		Protocol:              &url.Scheme,
		TopicArn:              aws.String(spec.ARN.String()),
		Attributes:            spec.SubscriptionAttributes,
		ReturnSubscriptionArn: aws.Bool(true),
	})

	switch {
	case isAWSError(err):
		// All documented API errors require some user intervention and
		// are not to be retried.
		// https://docs.aws.amazon.com/sns/latest/api/API_Subscribe.html#API_Subscribe_Errors
		status.MarkNotSubscribed(v1alpha1.AWSSNSReasonRejected, "Subscription request rejected")
		return controller.NewPermanentError(susbscribeErrorEvent(url, spec.ARN.String(), err))
	case err != nil:
		status.MarkNotSubscribed(v1alpha1.AWSSNSReasonFailedSync, "Cannot subscribe event source endpoint")
		return fmt.Errorf("%w", susbscribeErrorEvent(url, spec.ARN.String(), err))
	}

	logging.FromContext(ctx).Debug("Subscribe responded with: ", resp)

	status.MarkSubscribed()
	status.SubscriptionARN = resp.SubscriptionArn

	return reconciler.NewEvent(corev1.EventTypeNormal, ReasonSubscribed,
		"Subscribed to SNS topic %q", spec.ARN.String())
}

// ensureUnsubscribed ensures the source's HTTP(S) endpoint is unsubscribed
// from the SNS topic.
func (r *Reconciler) ensureUnsubscribed(ctx context.Context) error {
	src := v1alpha1.SourceFromContext(ctx)
	// TODO(antoineco): Follow up with a proper FindSubscription() method.
	// triggermesh/aws-event-sources#185
	subsARN := src.(*v1alpha1.AWSSNSSource).Status.SubscriptionARN

	// abandon if the subscription's ARN was never written to the source's status
	if subsARN == nil {
		return nil
	}

	spec := src.(apis.HasSpec).GetUntypedSpec().(v1alpha1.AWSSNSSourceSpec)

	snsClient, err := newSNSClient(r.secretsCli(src.GetNamespace()), spec.ARN.Region, &spec.Credentials)
	switch {
	case isNotFound(err):
		// the finalizer is unlikely to recover from a missing Secret,
		// so we simply record a warning event and return
		event.Warn(ctx, ReasonFailedUnsubscribe, "Secret missing while finalizing subscription %q. Ignoring: %s",
			*subsARN, err)
		return nil
	case err != nil:
		return fmt.Errorf("%w", reconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedUnsubscribe,
			"Error creating SNS client: %s", err))
	}

	resp, err := snsClient.UnsubscribeWithContext(ctx, &sns.UnsubscribeInput{
		SubscriptionArn: subsARN,
	})

	switch {
	case isNotFound(err):
		return reconciler.NewEvent(corev1.EventTypeNormal, ReasonUnsubscribed,
			"Subscription %q already absent, skipping finalization", *subsARN)
	case isDenied(err):
		// it is unlikely that we recover from validation errors in the
		// finalizer, so we simply record a warning event and return
		event.Warn(ctx, ReasonFailedUnsubscribe, "Authorization error finalizing subscription %q. Ignoring: %s",
			*subsARN, toErrMsg(err))
		return nil
	case err != nil:
		// wrap any other error to fail the finalization
		event := reconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedUnsubscribe,
			"Error finalizing event source %q: %s", *subsARN, toErrMsg(err))
		return fmt.Errorf("%w", event)
	}

	logging.FromContext(ctx).Debug("Unsubscribe responded with: ", resp)

	return reconciler.NewEvent(corev1.EventTypeNormal, ReasonUnsubscribed,
		"Subscription %q was successfully deleted", *subsARN)
}

// newSNSClient returns a new SNS client for the given region using static credentials.
func newSNSClient(cli coreclientv1.SecretInterface,
	region string, creds *v1alpha1.AWSSecurityCredentials) (*sns.SNS, error) {

	credsValue, err := awsCredentials(cli, creds)
	if err != nil {
		return nil, fmt.Errorf("reading AWS security credentials: %w", err)
	}

	cfg := session.Must(session.NewSession(aws.NewConfig().
		WithRegion(region).
		WithCredentials(credentials.NewStaticCredentialsFromCreds(*credsValue)),
	))

	return sns.New(cfg), nil
}

// awsCredentials returns the AWS security credentials referenced in the
// source's spec.
func awsCredentials(cli coreclientv1.SecretInterface,
	creds *v1alpha1.AWSSecurityCredentials) (*credentials.Value, error) {

	accessKeyID := creds.AccessKeyID.Value
	secretAccessKey := creds.SecretAccessKey.Value

	// cache a Secret object by name to avoid GET-ing the same Secret
	// object multiple times
	var secretCache map[string]*corev1.Secret

	if vfs := creds.AccessKeyID.ValueFromSecret; vfs != nil {
		secr, err := cli.Get(context.Background(), vfs.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// cache Secret containing the access key ID so it can be reused
		// below in case the same Secret contains the secret access key
		secretCache = map[string]*corev1.Secret{
			vfs.Name: secr,
		}

		accessKeyID = string(secr.Data[vfs.Key])
	}

	if vfs := creds.SecretAccessKey.ValueFromSecret; vfs != nil {
		var secr *corev1.Secret
		var err error

		if secretCache != nil && secretCache[vfs.Name] != nil {
			secr = secretCache[vfs.Name]
		} else {
			secr, err = cli.Get(context.Background(), vfs.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}

		secretAccessKey = string(secr.Data[vfs.Key])
	}

	return &credentials.Value{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	}, nil
}

// isNotFound returns whether the given error indicates that some resource was
// not found.
func isNotFound(err error) bool {
	if k8sErr := apierrors.APIStatus(nil); errors.As(err, &k8sErr) {
		return k8sErr.Status().Reason == metav1.StatusReasonNotFound
	}
	if awsErr := awserr.Error(nil); errors.As(err, &awsErr) {
		return awsErr.Code() == sns.ErrCodeNotFoundException
	}
	return false
}

// isDenied returns whether the given error indicates that a request to the SNS
// API could not be authorized.
func isDenied(err error) bool {
	if awsErr := awserr.Error(nil); errors.As(err, &awsErr) {
		return awsErr.Code() == sns.ErrCodeAuthorizationErrorException
	}
	return false
}

// isAWSError returns whether the given error is an AWS API error.
func isAWSError(err error) bool {
	awsErr := awserr.Error(nil)
	return errors.As(err, &awsErr)
}

// toErrMsg attempts to extract the message from the given error if it is an
// AWS error.
// Those errors are particularly verbose and include a unique request ID that
// causes an infinite loop of reconciliations when appended to a status
// condition. Some AWS errors are not recoverable without manual intervention
// (e.g. invalid secrets) so there is no point letting that behaviour happen.
func toErrMsg(err error) string {
	if awsErr := awserr.Error(nil); errors.As(err, &awsErr) {
		return awserr.SprintError(awsErr.Code(), awsErr.Message(), "", awsErr.OrigErr())
	}
	return err.Error()
}

// susbscribeErrorEvent returns a reconciler event indicating that an endpoint
// could not be subscribed to a SNS topic.
func susbscribeErrorEvent(url *apis.URL, topicARN string, origErr error) reconciler.Event {
	return reconciler.NewEvent(corev1.EventTypeWarning, ReasonFailedSubscribe,
		"Error subscribing endpoint %q to SNS topic %q: %s", url, topicARN, toErrMsg(origErr))
}
