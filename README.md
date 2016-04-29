# ECS Reference Architecture: Service Discovery
Service discovery is a key component of most distributed systems and service-oriented architectures. With service discovery, services are automatically discovered as they get created and terminated on a given infrastructure. This reference architecture illustrates how service discovery can be built on AWS.

## Background
Many AWS customers build service-oriented, distributed applications using services such as ECS or [Amazon EC2][2]. The distributed nature of this type of architecture requires a fair amount of integration and synchronization, and the answer to that problem is not trivial. Quite often, our customers build such a functionality themselves and this can be time-consuming. Or they use a third-party solution and this often comes with a financial cost.

## Solution
In this reference architecture, we propose that by leveraging Amazon ECS, [Amazon Route 53][3] and [AWS Lambda][4], we can eliminate a lot of the work required to install, operate, and scale Service Discovery at the cluster level.

In this example, a web portal application (PortalApp) presents information from a Twitch application (TwitchApp) and a GoodReads application (GoodreadsApp). As instances of these applications are created within ECS, they are placed behind [Elastic Load Balancing][5] load balancers. When they come up, they generate an event in [AWS CloudTrail][6] which is picked up by [Amazon CloudWatch Events][7]. This in turn triggers a Lambda function, which essentially "registers" the service into an Amazon Route 53 private hosted zone. That CNAME mapping then points to the appropriate load balancers. It is then what the web portal (PortalApp) uses to access both TwitchApp and GoodreadsApp.

Specifically, the architecture described in this [diagram][8] can be created with an [AWS CloudFormation][9] template. That [template][10] does the following:
- Creates a [VPC][11] with two subnets and their route tables, as well as an Internet gateway
- Creates appropriate [IAM][12] roles (for the container instances, ECS and Lambda)
- Deploys an ECS cluster onto which will be launched a web portal application, a Twitch application, and a GoodReads application
- Creates load balancers and [security groups][13] for the three applications
- Creates an [Auto Scaling group][14] for your ECS cluster, with its accompanying [launch configuration][15]
- Creates a private [Amazon Route 53 hosted zone][16] (i.e., internal DNS)

**Note:** The names of the ECS services have to be [DNS-compliant][17] as they are re-used as CNAME  values.

## Deploying the architecture
Here are the steps that you must take to deploy the architecture.

### Prerequisites
We expect that you have the following available:
- A host onto which you can build Docker images
	- it should have [Docker][17a] 
	- and [Git][17b] installed
	- as well as [AWS CLI][17c] installed and [configured][17d]. Make sure that the role that the AWS CLI will use is permissioned to push images to ECR (e.g. [AmazonEC2ContainerRegistryPowerUser][17e])

### Build the microservices app and push the images to [ECR][18a]

1. Use your desktop or an EC2 instance to build microservices container images. If you haven't installed Docker already, see the [documentation][19a] for further info.
2. Clone this repository. You should see a *microservices* directory with three sub-directories, each containing the information needed to build three Docker containers.
	```python
	> git clone https://github.com/awslabs/ecs-refarch-service-discovery
	```
3. Get the login credentials to ECR registry by typing below command
	```python
	> aws ecr get-login | sh
	```
4. Navigate to the [ECS Console][20b] and click on **Repositories** on the left. Create a new repository and specify the name '*twitchapp*'

5. Repeat the above step for '*goodreadsapp*' and '*portalapp*'

6. Build the Docker containers in each of the subdirectories:

```python
> cd microservices
> cd twitch
> docker build -t twitchapp .
> cd ../goodreads
> docker build -t goodreadsapp .
> cd ../portal
> docker build -t portalapp .
```

7. [Tag and Push the images](http://docs.aws.amazon.com/AmazonECR/latest/userguide/docker-push-ecr-image.html) to your ECR repository by typing these commands, replacing *123456789012* with your Account ID.

```python
> docker tag twitchapp:latest 123456789012.dkr.ecr.us-east-1.amazonaws.com/twitchapp:latest
> docker push 123456789012.dkr.ecr.us-east-1.amazonaws.com/twitchapp:latest
> docker tag goodreadsapp:latest 123456789012.dkr.ecr.us-east-1.amazonaws.com/goodreadsapp:latest
> docker push 123456789012.dkr.ecr.us-east-1.amazonaws.com/goodreadsapp:latest
> docker tag portalapp:latest 123456789012.dkr.ecr.us-east-1.amazonaws.com/portalapp:latest
> docker push 123456789012.dkr.ecr.us-east-1.amazonaws.com/portalapp:latest
```

> **Hint:** View the ECR repository in the ECS console and expand the **Build, tag, and push Docker image** section for more complete instructions.

![Build, tag, and push Docker image Console](https://s3.amazonaws.com/amazonecs-reference-architectures/service-discovery/build_tag_push.png)

### Launch the AWS CloudFormation template

1. Choose **Launch Stack** to launch the template in the us-east-1 region in your account:
[![Launch ECS Service Discovery into North Virginia with CloudFormation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/images/cloudformation-launch-stack-button.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/new?stackName=ecs-service-discovery&templateURL=https://s3.amazonaws.com/amazonecs-reference-architectures/service-discovery/ecs-refarch-service-discovery.template)
2. Give a Stack Name and select your preferred key name. If you do not have a key available, see [Amazon EC2 Key Pairs][22].
3. For each app - i.e. Twitch App, Goodreads App and Portal App - add the Docker image name - e.g. *123456789012.dkr.ecr.us-east-1.amazonaws.com/twitchapp:latest*
4. Choose **Next** and **Next**, check the acknowledgment, and choose **Create**.

This takes a few minutes; When CREATE_COMPLETE is displayed, note down the following values from the **Outputs** tab:
- Route53PrivateHostedZoneID
- LambdaServiceRole
- ECSClusterName

It might be a good idea to keep the **Output** tab open for the rest of this step-by-step guide.

### Create your [Lambda][4] function
1. Open the [Lambda console][23].
2. Choose **Create a Lambda function**, **Skip**.
3. Name your function, e.g. *registerEcsServiceDns*.
4. For **Runtime**, choose **Python 2.7** and replace the code with the content of this [Python file][24].
6. In the code, replace "Route53PrivateHostedZoneID" with your Route53PrivateHostedZoneID which was one of the output values from your AWS CloudFormation template.
7. In the code, replease "ECSClusterName" with your ECSClusterName which was one of the output values from your AWS CloudFormation template.
8. Under **Lambda function handler and role**, leave the default 'lambda_function.lambda_handler' for **Handler**. For **Role**, select the [service role][28] that was created you for by AWS CloudFormation. (You took note of it earlier.)
9. Increase the timeout to 10 sec and choose **Next**.
10. Review your Lambda function settings and choose **Create Function**.

### Create and configure the CloudWatch event that triggers your Lambda function
1. Open the [CloudWatch console][25].
2. On the left side, choose **Events**, **Create Rule**.
3. Choose **Show Advanced Options**, **Edit** to edit the JSON version. Replace the default value with the contents of the [linked CWE file][26].
5. Choose **Add Target** and select the Lambda function that was created in previous step.
6. Choose **Configure Details**, name your rule, and choose **Create Rule**.

### Create your three services

1. Open the [ECS console][27].
2. Select the cluster created for you by the AWS CloudFormation template.
3. Choose **Services**, **Create**.
4. For **Task Definition**, enter the ECS task definition created for you by the AWS CloudFormation template. You can find this in the *TwitchAppTaskDefinition* CloudFormation output. The CloudFormation output will be in the form of an ARN though so you want to only enter what is after "[...]:task-definition/". You will know that you have entered the right task definition when the Configure ELB button become active.
5. For **Service Name**, enter "TwitchApp".
6. For **Number of tasks**, enter "2".
7. Select **Configure ELB** to add a load balancer.
8. In **Select IAM role for service**, select the ECS service role created for you by the AWS CloudFormation template. You can find this in the *ECSServiceRole* CloudFormation output.
9. In **ELB Name**, select the ELB created for you by the AWS CloudFormation template. You can find this in the *LoadBalancerTwitchApp* CloudFormation output.
10. Choose **Save** and then **Create Service**.

Repeat this procedure for the other two services, modifying values as follows:
 - For GoodreadsApp:
   - Enter "GoodreadsApp" as **Service Name**.
   - Use the *GoodreadsAppTaskDefinition* CloudFormation output as **Task Definition**.
   - Use the *LoadBalancerGoodreadsApp* CloudFormation output as **ELB Name**.
 - For PortalApp:
   - Enter "PortalApp" as **Service Name**.
   - Use the *PortalAppTaskDefinition* CloudFormation output as **Task Definition**.
   - Use the *LoadBalancerPortalApp* CloudFormation output as **ELB Name**.

### Test your setup
1. Navigate to the details of your PortalApp service. (If you followed the steps above, that's the page that should be in front of you right now.)
2. Click the load balancer name to open the Elastic Load Balancing management page.
3. Copy the value in **DNS Name** and paste it into a browser's address bar.
4. Test GoodreadsApp: Enter and ISBN (e.g. "0316219282") and press enter.
5. Test TwitchApp: Enter a video game name (e.g. "Minecraft") and press enter.

## Conclusion
The goal of Service Discovery is essentially to allow for the components of a distributed architecture to find each other. This is achieved with two components:
- A location to centralize service information
- A mechanism to find and register those services in that location

In this example, we  used Amazon Route 53 and AWS Lambda, respectively: Amazon Route 53 acts as the repository for service registration, and adding and removing them is achieved via the Lambda function. This is made dynamic by triggering that function upon creation or deletion of services.

You can easily see the records that the Lambda function created for you. Just visit the [Route 53 Console][29], click on "Hosted zones" below DNS management and click on your specified hosted zone (default is "ecs.internal."). You will then see that three CNAME records were added pointing to each of the ELBs fronting the Portal, Goodreads and Twitch services.

In the setup, it is imperative for the web portal application to know the DNS name of the services in advance, which is why the instructions above were telling you to enter a specific service name ("TwitchApp", "GoodreadsApp", "PortalApp").

Note: Services behind load balancers render the infrastructure reliable; they can move around (as part of maintenance, or if the applications fail), with no impact to the service.

Feel free to re-use this example in your application, or take it apart and make it work for your context. For example, another place to store information about services could be an Amazon DynamoDB table, and another way to register services could be to have them self-register.

## Cleanup
To delete what you created:

1. Update service portalapp to 0 task, then delete it.
2. Do the same for both the goodreadsapp and the TwitchApp.
3. Delete the Lambda function.
4. Delete the CloudWatch Events rule.
5. Delete the ECR Repositories - twitchapp, goodreadsapp, portalapp
6. Delete the AWS CloudFormation template.

## License
This reference architecture sample is licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0.


[1]: http://aws.amazon.com/ecs/
[2]: http://aws.amazon.com/ec2/
[3]: http://aws.amazon.com/route53/
[4]: http://aws.amazon.com/lambda/
[5]: https://aws.amazon.com/elasticloadbalancing/
[6]: https://aws.amazon.com/cloudtrail/
[7]: http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/WhatIsCloudWatchEvents.html
[8]: https://s3.amazonaws.com/amazonecs-reference-architectures/service-discovery/ecs-refarch-service-discovery.pdf
[9]: https://aws.amazon.com/cloudformation/
[10]: cfn-templates/ecs-refarch-service-discovery.template
[11]: https://aws.amazon.com/vpc/
[12]: https://aws.amazon.com/iam/
[13]: http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Security.html
[14]: https://aws.amazon.com/autoscaling/
[15]: http://docs.aws.amazon.com/AutoScaling/latest/DeveloperGuide/LaunchConfiguration.html
[16]: http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/AboutHostedZones.html
[17]: https://tools.ietf.org/html/rfc1035
[17a]: https://docs.docker.com/engine/installation/
[17b]: https://git-scm.com/book/en/v2/Getting-Started-Installing-Git
[17c]: https://docs.aws.amazon.com/cli/latest/userguide/installing.html
[17d]: https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html
[17e]: https://docs.aws.amazon.com/AmazonECR/latest/userguide/ecr_managed_policies.html#AmazonEC2ContainerRegistryPowerUser
[18a]: http://aws.amazon.com/ecr/
[19a]: http://docs.aws.amazon.com/AmazonECS/latest/developerguide/docker-basics.html#install_docker
[20]: https://console.aws.amazon.com/cloudformation
[20a]: https://github.com/awslabs/service-discovery-ecs-consul
[20b]: https://console.aws.amazon.com/ecs/home
[21a]: http://docs.aws.amazon.com/IAM/latest/UserGuide/id_users.html
[22a]: http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html
[23a]: http://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies.html
[21]: https://aws.amazon.com/s3/
[22]: http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html
[23]: https://console.aws.amazon.com/lambda/
[24]: ecs-register-service-dns-lambda.py
[25]: https://console.aws.amazon.com/cloudwatch
[26]: cwe-ecs-rule.json
[27]: https://console.aws.amazon.com/ecs/
[28]: http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-service.html
[29]: https://console.aws.amazon.com/route53/home
