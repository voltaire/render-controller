import iam = require('@aws-cdk/aws-iam');
import route53 = require('@aws-cdk/aws-route53');
import s3 = require('@aws-cdk/aws-s3');
import sns = require('@aws-cdk/aws-sns');
import subscriptions = require('@aws-cdk/aws-sns-subscriptions');
import * as cdk from '@aws-cdk/core';

export class SkskskStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props)

    const sepAccount = new iam.AccountPrincipal('006851364659')

    const deployGroup = new iam.Group(this, 'sksksk-deploy', {})

    deployGroup.addToPolicy(new iam.PolicyStatement({
      actions: [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "cloudwatch:PutMetricData",
        "cloudwatch:ListMetrics",
        "cloudwatch:GetMetricData",
      ],
      resources: ["*"],
    }))

    const deployUser = new iam.User(this, 'deployUser', {
      groups: [deployGroup],
    })

    const tonkatsuZone = route53.HostedZone.fromHostedZoneAttributes(this, 'tonkatsuZone', {
      hostedZoneId: 'ZVAMW53PNR70P',
      zoneName: 'tonkat.su',
    })

    const sepBucket = s3.Bucket.fromBucketName(this, 'sepBucket', 'mc.sep.gg-backups')
    sepBucket.grantRead(deployGroup)

    const backupNotificationTopic = new sns.Topic(this, "SkskskBackupTopic", {});
    backupNotificationTopic.grantPublish(sepAccount)
    backupNotificationTopic.addToResourcePolicy(new iam.PolicyStatement({
      actions: ["sns:Publish"],
      resources: [backupNotificationTopic.topicArn],
      principals: [new iam.ServicePrincipal("s3.amazonaws.com")],
      conditions: {
        ArnEquals: {"aws:SourceArn": sepBucket.bucketArn}
      },
    }))
    deployGroup.addToPolicy(new iam.PolicyStatement({
      actions: ["sns:Subscribe", "sns:ConfirmSubscription"],
      resources: [backupNotificationTopic.topicArn],
    }))

    new route53.CnameRecord(this, 'mapCname', {
      zone: tonkatsuZone,
      recordName: 'map',
      domainName: 'map.tonkat.su.website-us-east-1.linodeobjects.com',
    })

    new route53.CnameRecord(this, 'oldmapCname', {
      zone: tonkatsuZone,
      recordName: 'oldmap',
      domainName: 'oldmap.tonkat.su.website-us-east-1.linodeobjects.com',
    })

    new route53.CnameRecord(this, 'bungeeCord', {
      zone: tonkatsuZone,
      recordName: 'mc',
      domainName: 'mc.sep.gg',
    })

    const rendererRecord = new route53.ARecord(this, 'renderer', {
      zone: tonkatsuZone,
      recordName: 'renderer',
      target: route53.RecordTarget.fromIpAddresses('107.150.36.10'),
    })

    backupNotificationTopic.addSubscription(new subscriptions.UrlSubscription('https://'+rendererRecord.domainName+'/callback', {protocol: sns.SubscriptionProtocol.HTTPS}))
  }
}