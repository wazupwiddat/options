#!/bin/bash

# Define variables
CLUSTER_NAME="OptionsAPIStack-ECSCluster-47tv7oa5GeBu"
SERVICE_NAME="OptionsAPIService"
TASK_FAMILY="options-api-task"
CONTAINER_NAME="options-api-container"
NEW_IMAGE="026450499422.dkr.ecr.us-east-1.amazonaws.com/options:latest"

# Fetch the latest ACTIVE task definition ARN and definition
TASK_DEFINITION=$(aws ecs describe-task-definition --task-definition $TASK_FAMILY --query 'taskDefinition.taskDefinitionArn' --output text)

# Create a new task definition revision with the new image
NEW_TASK_DEF=$(aws ecs register-task-definition \
    --family $(aws ecs describe-task-definition --task-definition $TASK_DEFINITION --query 'taskDefinition.family' --output text) \
    --container-definitions "[{\"name\": \"$CONTAINER_NAME\", \"image\": \"$NEW_IMAGE\", \"cpu\": 0, \"memory\": 256, \"essential\": true, \"portMappings\": [{\"containerPort\": 8080, \"hostPort\": 8080}]}]" \
    --requires-compatibilities $(aws ecs describe-task-definition --task-definition $TASK_DEFINITION --query 'taskDefinition.requiresCompatibilities' --output text) \
    --network-mode $(aws ecs describe-task-definition --task-definition $TASK_DEFINITION --query 'taskDefinition.networkMode' --output text) \
    --cpu $(aws ecs describe-task-definition --task-definition $TASK_DEFINITION --query 'taskDefinition.cpu' --output text) \
    --memory $(aws ecs describe-task-definition --task-definition $TASK_DEFINITION --query 'taskDefinition.memory' --output text) \
    --execution-role-arn $(aws ecs describe-task-definition --task-definition $TASK_DEFINITION --query 'taskDefinition.executionRoleArn' --output text) \
    --query 'taskDefinition.taskDefinitionArn' --output text)

# Update ECS service
aws ecs update-service --cluster $CLUSTER_NAME --service $SERVICE_NAME --task-definition $NEW_TASK_DEF

echo "Service updated to use new image: $NEW_IMAGE"