package email

import (
	"context"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/logger"
	"the-slot-finder/server/internal/monitoring"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"

	"github.com/getsentry/sentry-go"
)

type SESSender struct {
	client *ses.Client
	config config.Config
}

func NewSESSender(config config.Config) (*SESSender, error) {
	// Load AWS config
	cfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRegion(config.AWSSESRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AWSSESAccessKey, config.AWSSESSecretKey, "")),
	)
	if err != nil {
		if hub := monitoring.GetSentryHub(context.TODO()); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("service", "ses")
				scope.SetTag("operation", "send_email")
				monitoring.SetScopeData(scope, "step", "error")
				monitoring.SetScopeData(scope, "error_details", err.Error())
				hub.CaptureException(errors.ExternalServiceError("failed to load AWS config", err))
			})
		}

		logger.Sugar.Errorw("Failed to load AWS config",
			"service", "ses",
			"operation", "send_email",
			"region", config.AWSSESRegion,
			"error", err.Error(),
		)

		return nil, errors.ExternalServiceError("Failed to load AWS config", err).
			WithOperation("load_aws_config").
			WithResource("ses")
	}

	// Create SES client
	client := ses.NewFromConfig(cfg)

	return &SESSender{
		client: client,
		config: config,
	}, nil
}

func (s *SESSender) SendEmail(ctx context.Context, request EmailRequest) (*EmailResponse, error) {
	// Build the email content
	var body *types.Body

	switch {
	case request.HTMLBody != "":
		body = &types.Body{
			Html: &types.Content{
				Data:    aws.String(request.HTMLBody),
				Charset: aws.String("UTF-8"),
			},
		}
		if request.TextBody != "" {
			body.Text = &types.Content{
				Data:    aws.String(request.TextBody),
				Charset: aws.String("UTF-8"),
			}
		}
	case request.TextBody != "":
		body = &types.Body{
			Text: &types.Content{
				Data:    aws.String(request.TextBody),
				Charset: aws.String("UTF-8"),
			},
		}
	default:
		return nil, errors.ExternalServiceError("either HTML or text content must be provided", nil).
			WithOperation("send_email").
			WithResource("ses")
	}

	// Build the message
	message := &types.Message{
		Subject: &types.Content{
			Data:    aws.String(request.Subject),
			Charset: aws.String("UTF-8"),
		},
		Body: body,
	}

	// Build the destination
	destination := &types.Destination{
		ToAddresses: request.To,
	}
	if len(request.Cc) > 0 {
		destination.CcAddresses = request.Cc
	}
	if len(request.Bcc) > 0 {
		destination.BccAddresses = request.Bcc
	}

	// Build the input
	input := &ses.SendEmailInput{
		Source:      aws.String(s.config.AWSSESFromEmail),
		Destination: destination,
		Message:     message,
	}

	// Send the email
	result, err := s.client.SendEmail(ctx, input)
	if err != nil {
		if hub := monitoring.GetSentryHub(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("service", "ses")
				scope.SetTag("operation", "send_email")
				monitoring.SetScopeData(scope, "step", "error")
				monitoring.SetScopeData(scope, "error_details", err.Error())
				monitoring.SetScopeData(scope, "recipients", len(request.To))
				hub.CaptureException(errors.ExternalServiceError("failed to send email via SES", err))
			})
		}

		logger.Sugar.Errorw("Failed to send email via SES",
			"service", "ses",
			"operation", "send_email",
			"recipients", len(request.To),
			"error", err.Error(),
		)

		return &EmailResponse{
			Provider: "ses",
			Status:   "failed",
			Error:    err.Error(),
		}, errors.ExternalServiceError("Failed to send email via SES", err).
			WithOperation("send_email").
			WithResource("ses")
	}

	return &EmailResponse{
		MessageID: *result.MessageId,
		Provider:  "ses",
		Status:    "sent",
	}, nil
}

// SendRawEmail sends a raw email (useful for complex email structures)
func (s *SESSender) SendRawEmail(ctx context.Context, rawData []byte) (*EmailResponse, error) {
	input := &ses.SendRawEmailInput{
		RawMessage: &types.RawMessage{
			Data: rawData,
		},
	}

	result, err := s.client.SendRawEmail(ctx, input)
	if err != nil {
		return &EmailResponse{
			Provider: "ses",
			Status:   "failed",
			Error:    err.Error(),
		}, errors.ExternalServiceError("Failed to send raw email via SES", err).
			WithOperation("send_raw_email").
			WithResource("ses")
	}

	return &EmailResponse{
		MessageID: *result.MessageId,
		Provider:  "ses",
		Status:    "sent",
	}, nil
}
