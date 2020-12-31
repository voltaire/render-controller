#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { SkskskStack } from '../lib/sksksk-stack';

const app = new cdk.App();
new SkskskStack(app, 'SkskskStack', {
    env: {
        region: 'us-west-2',
    }
});
