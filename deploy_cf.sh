#!/bin/bash

# Variables
TEMPLATE_FILE="cloudformation.yaml"
STACK_NAME="OptionsAPIStack"
BUCKET_NAME="jdub-option-images"  # Update this with your actual S3 bucket name
REGION="us-east-1"
ECR_IMAGE_URI="026450499422.dkr.ecr.us-east-1.amazonaws.com/options:latest"  # Update this if needed
CERT_ARN="arn:aws:acm:us-east-1:026450499422:certificate/390e9185-4df5-4405-89d3-38ec9e090bac"

# Validate the CloudFormation template
echo "Validating the CloudFormation template..."
aws cloudformation validate-template --template-body file://$TEMPLATE_FILE
if [ $? -ne 0 ]; then
    echo "Template validation failed."
    exit 1
else
    echo "Template successfully validated."
fi

# Upload the CloudFormation template to S3
echo "Uploading template to S3..."
aws s3 cp $TEMPLATE_FILE s3://$BUCKET_NAME/$TEMPLATE_FILE
if [ $? -ne 0 ]; then
    echo "Failed to upload template to S3."
    exit 1
else
    echo "Template uploaded successfully."
fi

# Check if the stack already exists
aws cloudformation describe-stacks --stack-name $STACK_NAME > /dev/null 2>&1

if [ $? -eq 0 ]; then
    # Update the stack
    echo "Updating CloudFormation stack..."
    aws cloudformation update-stack --stack-name $STACK_NAME --template-url https://$BUCKET_NAME.s3.$REGION.amazonaws.com/$TEMPLATE_FILE --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --parameters ParameterKey=ECRImageURI,ParameterValue=$ECR_IMAGE_URI ParameterKey=CertificateArn,ParameterValue=$CERT_ARN
    if [ $? -ne 0 ]; then
        echo "Failed to update stack."
        exit 1
    else
        echo "Stack update initiated successfully."
    fi
else
    # Create the stack
    echo "Creating CloudFormation stack..."
    aws cloudformation create-stack --stack-name $STACK_NAME --template-url https://$BUCKET_NAME.s3.$REGION.amazonaws.com/$TEMPLATE_FILE --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND --parameters ParameterKey=ECRImageURI,ParameterValue=$ECR_IMAGE_URI ParameterKey=CertificateArn,ParameterValue=$CERT_ARN
    if [ $? -ne 0 ]; then
        echo "Failed to create stack."
        exit 1
    else
        echo "Stack creation initiated successfully."
    fi
fi

# Wait for the stack to be created or updated
echo "Waiting for stack operation to complete..."
aws cloudformation wait stack-create-complete --stack-name $STACK_NAME
if [ $? -ne 0 ]; then
    echo "Stack operation did not complete successfully."
    exit 1
else
    echo "Stack operation completed successfully."
fi
