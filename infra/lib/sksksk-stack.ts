import * as assets from '@aws-cdk/aws-s3-assets';
import * as events from '@aws-cdk/aws-events';
import * as events_targets from '@aws-cdk/aws-events-targets';
import * as iam from '@aws-cdk/aws-iam';
import * as lambda from '@aws-cdk/aws-lambda';
import * as logs from '@aws-cdk/aws-logs';
import * as route53 from '@aws-cdk/aws-route53';
import * as s3 from '@aws-cdk/aws-s3';
import * as sns from '@aws-cdk/aws-sns';
import * as subscriptions from '@aws-cdk/aws-sns-subscriptions';
import * as path from 'path';
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

    const dataBucket = new s3.Bucket(this, 'dataBucket', {
      accessControl: s3.BucketAccessControl.PRIVATE,
      autoDeleteObjects: true,
    })

    const lambdasAsset = new assets.Asset(this, 'lambdasZip', {
      path: path.join(__dirname, '../../build/'),
    })

    const updateMapCertLambda = new lambda.Function(this, 'updateMapCertLambda', {
      code: lambda.Code.fromBucket(
        lambdasAsset.bucket,
        lambdasAsset.s3ObjectKey,
      ),
      runtime: lambda.Runtime.GO_1_X,
      handler: 'update-map-cert',
      logRetention: logs.RetentionDays.THREE_DAYS,
      environment: {
        DATA_BUCKET_NAME: dataBucket.bucketName,
        RENEW_IF_WITHIN: '336h', // 14 days
      },
    })

    dataBucket.grantReadWrite(updateMapCertLambda)

    new events.Rule(this, 'mapCertCron', {
      schedule: events.Schedule.rate(cdk.Duration.days(7)),
      targets: [new events_targets.LambdaFunction(updateMapCertLambda)],
    })
  }
}