# ECS Reference Architecture: Service Discovery
Service discovery is automatically detecting services as they live and die on a given infrastructure. The Service Discovery for [Amazon ECS][1] Reference Architecture illustrates how service discovery can be built on AWS. 

## Background
Many AWS customers build service-oriented, distributed applications using services such as ECS or [Amazon EC2][2]. The distributed nature of this type of architecture requires a fair amount of integration and synchronization, and the answer to that problem is not trivial. Quite often, our customers build such a functionality themselves (often time-consuming), or they use a third-party (often pricey) solution.

## Solution
In this Reference Architecture, we propose that by leveraging ECS, [Amazon Route 53][3] and [AWS Lambda][4], we can eliminate a lot of the work required to install, operate, and scale a cluster management infrastructure. 

In this example, a web portal application (PortalApp) presents information from a stock application (StockApp) and a weather application (WeatherApp). As instances of these applications are created within ECS, they are placed behind [Elastic Load Balancing][5] load balancers. When they come up, they generate an event in [AWS CloudTrail][6] which is picked up by [Amazon CloudWatch Events][7]. This in turn triggers a Lambda function, which essentially "registers" the service into an Amazon Route 53 private hosted zone. That CNAME mapping then points to the appropriate load balancers. It is also what the web portal uses in this case to access both StockApp and WeatherApp.

Specifically, the architecture described in this [diagram][8] can be created with an [AWS CloudFormation][9] template. That [template][10] does the following:
- Creates a [VPC][11] with two subnets and their route tables, as well as an Internet gateway
- Creates appropriate [IAM][12] roles
- Deploys an ECS cluster onto which will be launched a web portal application, a stock application, and a weather application
- Creates [load balancers][5] and [security groups][13] for the three applications
- Creates an [Auto Scaling group][14] for your ECS cluster, with its accompanying [launch configuration][15]
- Creates a private [Amazon Route 53 hosted zone][16] (i.e., internal DNS)

Note: The name of the ECS services have to be [DNS-compliant][17] as they are re-used as CNAME  values.

## Deploying the architecture
Here are the steps that need to be taken to deploy the architecture.

### Build the microservices app and push the images to [ECR][18a]
1. Use your desktop or an EC2 instance to build microservices container images. If you haven't installed docker already, see the [documentation][19a] for further info. 
2. Download the [source code for the three microservices][20a]. You should see three directories that contain the information needed to build three Docker containers. 
3. Get the login credentials to ECR registry by typing below command and authenticate your Docker client by running the **docker login command**
```python
> sudo aws ecr get-login
```
4. Navigate to ECS Console and click on **Repositories** on the left. Create three new repositories by specifying a name 'weatherapp', 'stockapp' and 'portalapp'
5. Ensure your [IAM user][21a] or [IAM roles][22a] is attached to `AmazonEC2ContainerRegistryPowerUser` [IAM policy][23a]. This will give the user or role access to create the container Image in your repository
6. Build the Docker containers in each of the subdirectories: 
```python
> $ cd weather
> $ sudo docker build -t weatherapp .
> $ cd ../stock-price
> $ sudo docker build -t stockapp .
> $ cd ../portal
> $ sudo docker build -t portalapp .
```
7. Tag and Push the images to your ECR repository by typing these commands, replacing 101010101010 with your Account ID **HINT** You will find complete instruction in each repository and by clicking small triangle
```python
> sudo docker tag weatherapp:latest 101010101010.dkr.ecr.us-east-1.amazonaws.com/weatherapp:latest
> sudo docker push 101010101010.dkr.ecr.us-east-1.amazonaws.com/weatherapp:latest
> sudo docker tag stockapp:latest 101010101010.dkr.ecr.us-east-1.amazonaws.com/stockapp:latest
> sudo docker push 101010101010.dkr.ecr.us-east-1.amazonaws.com/stockapp:latest
> sudo docker tag portalapp:latest 101010101010.dkr.ecr.us-east-1.amazonaws.com/portalapp:latest
> sudo docker push 101010101010.dkr.ecr.us-east-1.amazonaws.com/portalapp:latest
```

### Get an API key from OpenWeatherMap
1. Navigate to [OpenWeatherMap][18] and create an account.
2. The API key is available from the [OpenWeatherMap home page][19].

### Launch the AWS CloudFormation template

1. Choose **Launch Stack** to launch the template in the us-east-1 region in your account:
[![Launch ECS Service Discovery into North Virginia with CloudFormation](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/images/cloudformation-launch-stack-button.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/new?stackName=ecs-service-discovery&templateURL=https://s3.amazonaws.com/ecs-service-discovery-reference-architecture/ecs-refarch-service-discovery.template)
2. Select your preferred key name. If you do not have a key available, see [Amazon EC2 Key Pairs][22].
3. Fill in your Weather API key.
4. Choose **Next** and **Next**, check the acknowledgment, and choose **Create**.

This takes a few minutes; When CREATE_COMPLETE is displayed, note down the following values from the **Outputs** tab:
- Route53PrivateHostedZoneID
- LambdaServiceRole
- ECSClusterName

It might be a good idea to keep the **Output** tab open for the rest of this step-by-step guide.

### Create your [Lambda][4] function
1. Open the [Lambda console][23].
2. Choose **Create a Lambda function**, **Skip**.
3. Name your function. 
4. For **Runtime**, choose **Python** and replace the code with the content of this [Python file][24].
6. In the code, replace "Route53PrivateHostedZoneID" with your Route53PrivateHostedZoneID which was one of the output values from your AWS CloudFormation template.
7. In the code, replease "ECSClusterName" with your ECSClusterName which was one of the output values from your AWS CloudFormation template.
8. Under **Lambda function handler and role**, leave the default 'lambda_function.lambda_handler' for **Handler**. For **Role**, select the [service role][28] that was created you for by AWS CloudFormation. (You took note of it earlier.)
9. Increase the timeout to 10 sec and choose **Next**.
10. Review your Lambda function settings and choose **Create Function**.

### Create and configure the [CloudWatch event][7] that triggers your Lambda function
1. Open the [CloudWatch console][25].
2. On the left side, choose **Events**, **Create Rule**.
3. Choose **Show Advanced Options**, **Edit** to edit the JSON version. Replace the default value with the contents of the [linked CWE file][26].
5. Choose **Add Target** and select your Lambda function.
6. Choose **Configure Details**, name your rule, and choose **Create Rule**.

### Create your three services

1. Open the [ECS console][27].
2. Select the cluster created for you by the AWS CloudFormation template.
3. Choose **Services**, **Create**.
4. For **Service Name**, enter *weatherapp* for the WeatherApp task definition.
5. For **Number of tasks**, enter "2".
6. Select the load balancer created for you that ends with ":8080".
7. Select the ECS service role created for you by the AWS CloudFormation template.
8. Choose **Create Service**.

Repeat this procedure for the other two services, modifying values as follows:
 - For **Service Name**, enter *stockapp* for the StockApp task definition. Select the load balancer created for you that ends with ":9090".
 - For **Service Name**, enter *portalapp* for the PortalApp task definition. Select the load balancer created for you that ends with ":80".

### Test your setup
1. Navigate to the details of your portalapp service. (If you followed the steps above, that's the page that should be in front of you right now.)
2. Choose the load balancer name to open the Elastic Load Balancing management page.
3. Copy the value in **DNS Name** and paste it into a browser.
4. Enter "AMZN" in the StocksApp, and choose **Add**.
5. Enter "New York City" in the WeatherApp, and choose **Add**.

## Conclusion
The goal of Service Discovery is essentially to allow for the components of a distributed architecture to find each other. This is achieved with two components:
- A location to centralize service information
- A mechanism to find and register those services in that location 

In this example, we  used Amazon Route 53 and AWS Lambda, respectively: Amazon Route 53 acts as the repository for service registration, and adding and removing them is achieved via the Lambda function. This is made dynamic by triggering that function upon creation or deletion of services.

In the setup, it is imperative for the web portal application to know the DNS name of the services in advance, which is why the instructions above were instructing you to enter a specific service name ("WeatherApp", "StocksApp", "PortalApp").

Note: Services behind load balancers render the infrastructure reliable; they can move around (as part of maintenance, or if the applications fail), with no impact to the service.

Feel free to re-use this example in your application, or take it apart and make it work for your context. For example, another place to store information about services could be an Amazon DynamoDB table, and another way to register services could be to have them self-register.

## Cleanup
To delete what you created:
1. Update service portalapp to 0 task, then delete it.
2. Do the same for both StocksApp and the WeatherApp.
3. Delete the Lambda function.
4. Delete the CloudWatch Events rule.
5. Delete the AWS CloudFormation template.

## License
This reference architecture sample is licensed under Apache 2.0.
 

[1]: http://aws.amazon.com/ecs/
[2]: http://aws.amazon.com/ec2/ 
[3]: http://aws.amazon.com/route53/
[4]: http://aws.amazon.com/lambda/
[5]: https://aws.amazon.com/elasticloadbalancing/
[6]: https://aws.amazon.com/cloudtrail/
[7]: http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/WhatIsCloudWatchEvents.html
[8]: <INSERT LINK TO DIAGRAM OF ARCHITECTURE>
[9]: https://aws.amazon.com/cloudformation/
[10]: cfn-templates/ecs-refarch-service-discovery.template
[11]: https://aws.amazon.com/vpc/
[12]: https://aws.amazon.com/iam/
[13]: http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Security.html
[14]: https://aws.amazon.com/autoscaling/
[15]: http://docs.aws.amazon.com/AutoScaling/latest/DeveloperGuide/LaunchConfiguration.html
[16]: http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/AboutHostedZones.html
[17]: https://tools.ietf.org/html/rfc1035
[18a]: http://aws.amazon.com/ecr/
[19a]: http://docs.aws.amazon.com/AmazonECS/latest/developerguide/docker-basics.html#install_docker
[20a]: https://github.com/awslabs/service-discovery-ecs-consul
[21a]: http://docs.aws.amazon.com/IAM/latest/UserGuide/id_users.html
[22a]: http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html
[23a]: http://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies.html

[18]: http://openweathermap.org/
[19]: http://home.openweathermap.org/
[20]: https://console.aws.amazon.com/cloudformation
[21]: https://aws.amazon.com/s3/
[22]: http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html
[23]: https://console.aws.amazon.com/lambda/
[24]: ecs-register-service-dns-lambda.py
[25]: https://console.aws.amazon.com/cloudwatch
[26]: cwe-ecs-rule.json
[27]: https://console.aws.amazon.com/ecs/	
[28]: http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-service.html

