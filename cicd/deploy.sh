#!/bin/sh
myS3BucketForDeployment="s3://edfx-rbn-pileface/swagger/"

currentDir=`pwd`

aws s3 cp --profile edfx-poc ../apiDefinition/pileface-go.yaml $myS3BucketForDeployment

sam build
sam deploy
