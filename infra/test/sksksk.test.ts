import { expect as expectCDK, matchTemplate, MatchStyle } from '@aws-cdk/assert';
import * as cdk from '@aws-cdk/core';
import * as Sksksk from '../lib/sksksk-stack';

test('Empty Stack', () => {
    const app = new cdk.App();
    // WHEN
    const stack = new Sksksk.SkskskStack(app, 'MyTestStack');
    // THEN
    expectCDK(stack).to(matchTemplate({
      "Resources": {}
    }, MatchStyle.EXACT))
});
