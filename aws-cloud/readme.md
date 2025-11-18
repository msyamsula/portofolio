# pre requisite setup
- install aws cli
- configure: `aws configure`
- install istio: `brew install istioctl`
- run aws services: `aws cloudformation deploy --template-file <YAML> --stack-name <NAME> --parameter-overrides DBPassword='<PASSWORD>' YourIP='<IP>'`
- delete: `aws cloudformation delete-stack --stack-name <NAME>`


# dummy program
- callee: `docker build --platform linux/amd64,linux/arm64 -t callee -f aws-cloud/services/callee/dockerfile .`
- caller: `docker build --platform linux/amd64,linux/arm64 -t caller -f aws-cloud/services/caller/dockerfile .`
- run: use docker desktop for easy to use, run testing image

# setup ecr (aws docker repository)
- get account id: `aws sts get-caller-identity`
- login to ecr: `aws ecr get-login-password --region <REGION> | docker login --username AWS --password-stdin <ID>.dkr.ecr.<REGION>.amazonaws.com`
- proper tagging: `docker tag <YOUR_IMAGE> <ACCOUNT_ID>.dkr.ecr.<REGION>.amazonaws.com/my-app:latest`
- push: `docker push <ACCOUNT_ID>.dkr.ecr.<REGION>.amazonaws.com/my-app:latest`
- pull: `docker pull <ACCOUNT_ID>.dkr.ecr.<REGION>.amazonaws.com/my-app:latest`

# setup eks
- add cluster to kubectl context for discovery: `aws eks --region <REGION> update-kubeconfig --name <CLUSTER>`
- run eks: `aws cloudformation deploy --template-file <FILE> --stack-name <NAME> --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM`



