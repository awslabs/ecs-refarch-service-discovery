from __future__ import print_function

import json
import boto3

def lambda_handler(event, context):
  
  # spit out event data
	print("Received event: " + json.dumps(event, indent=2))    

	# private hosted zone domain name and id
	privatezone = 'ecs.internal'
	zoneid = 'Route53PrivateHostedZoneID'
	cluster = 'ECSClusterName'

	# grab load balancer and service names
	lb = event['detail']['responseElements']['service']['loadBalancers'][0]['loadBalancerName']
	service = event['detail']['responseElements']['service']['serviceName']

	# check we are working against the appropriate ecs cluster
	if cluster != event['detail']['requestParameters']['cluster']:

		print("This event does not apply to us. No action taken.")
		return 0

	# grab DNS name for load balancer
	elbclient = boto3.client('elb')
	describelb = elbclient.describe_load_balancers(
		LoadBalancerNames=[
			lb
		]
	)
	lbcanonical = describelb['LoadBalancerDescriptions'][0]['DNSName']
	servicerecord = service + "." + privatezone + "."

	# grab type of event
	eventname = event['detail']['eventName']

	# boto connect to route53
	route53client = boto3.client('route53')

	# create/update record
	if eventname == 'CreateService':

		response = route53client.change_resource_record_sets(
			HostedZoneId=zoneid,
			ChangeBatch={
				'Comment' : 'ECS service registered',
				'Changes' : [
					{
						'Action' : 'UPSERT',
						'ResourceRecordSet' : {
							'Name' : servicerecord,
							'Type' : 'CNAME',
							'TTL' : 60,
							'ResourceRecords' : [
								{
									'Value' : lbcanonical
								}
							]
						}
					}
				] 
			}
		)

		print(response)
		return response

	# delete record
	elif eventname == 'DeleteService':

		response = route53client.change_resource_record_sets(
			HostedZoneId=zoneid,
			ChangeBatch={
				'Comment' : 'ECS service deregistered',
				'Changes' : [
					{
						'Action' : 'DELETE',
						'ResourceRecordSet' : {
							'Name' : servicerecord,
							'Type' : 'CNAME',
							'TTL' : 60,
							'ResourceRecords' : [
								{
									'Value' : lbcanonical
								}
							]
						}
					}
				] 
			}
		)

		print(response)
		return response

	else:

			print("This event does not apply to us. No action taken.")
			return 0
