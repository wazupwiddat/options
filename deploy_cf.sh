#!/bin/bash

# Variables
TEMPLATE_FILE="cloudformation.yaml"
STACK_NAME="GoAPIStack"
BUCKET_NAME="your-s3-bucket-name"
REGION="us-east-1"
PARAMETERS="ParameterKey=SubnetId,ParameterValue=subnet-abc123 ParameterKey=SecurityGroupId,ParameterValue=sg-abc123 ParameterKey=ECRImageURI,ParameterValue=026450499422.dkr.ecr.us-east-1.amazonaws.com/options:latest"

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
    aws cloudformation update-stack --stack-name $STACK_NAME --template-url https://$BUCKET_NAME.s3.amazonaws.com/$TEMPLATE_FILE --parameters $PARAMETERS --capabilities "CAPABILITY_IAM" "CAPABILITY_NAMED_IAM" "CAPABILITY_AUTO_EXPAND"
    if [ $? -ne 0 ]; then
        echo "Failed to update stack."
        exit 1
    else
        echo "Stack update initiated successfully."
    fi
else
    # Create the stack
    echo "Creating CloudFormation stack..."
    aws cloudformation create-stack --stack-name $STACK_NAME --template-url https://$BUCKET_NAME.s3.amazonaws.com/$TEMPLATE_FILE --parameters $PARAMETERS --capabilities "CAPABILITY_IAM" "CAPABILITY_NAMED_IAM" "CAPABILITY_AUTO_EXPAND"
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
