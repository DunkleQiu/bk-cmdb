/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logics

import (
	"configcenter/src/common/blog"
	"configcenter/src/common/http/rest"
	"configcenter/src/common/metadata"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	tcCommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcRegions "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	tcVpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

func (lgc *Logics) GetAwsRegions(kit *rest.Kit, secretID, secretKey string) ([]string, error) {
	sess, err := lgc.AwsNewSession(kit, "", secretID, secretKey)
	if err != nil {
		blog.ErrorJSON("getAwsRegions get aws new session failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	ec2Svc := ec2.New(sess)
	rsp, err := ec2Svc.DescribeRegions(nil)
	if err != nil {
		blog.ErrorJSON("getAwsRegions, sdk api DescribeRegions failed, err: %v. rid: %s", err, kit.Rid)
		return nil, err
	}

	regions := make([]string, 0)
	for _, region := range rsp.Regions {
		regions = append(regions, *region.RegionName)
	}

	return regions, nil
}

func (lgc *Logics) GetTencentCloudRegions(kit *rest.Kit, secretID, secretKey string) ([]string, error) {
	credential := lgc.TencentCloudNewCredential(kit, secretID, secretKey)

	client, err := cvm.NewClient(credential, tcRegions.Guangzhou, profile.NewClientProfile())
	if err != nil {
		blog.ErrorJSON("getTencentCloudRegions new client failed, err: %v, rid: %s", err, kit.Rid)
		return nil, nil
	}

	regionRequest := cvm.NewDescribeRegionsRequest()
	rsp, err := client.DescribeRegions(regionRequest)
	regions := make([]string, 0)
	for _, region := range rsp.Response.RegionSet {
		regions = append(regions, *region.Region)
	}

	return regions, nil
}

func (lgc *Logics) AwsNewSession(kit *rest.Kit, region, secretID, secretKey string) (*session.Session, error) {
	if region == "" {
		region = "us-west-2"
	}
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(secretID, secretKey, ""),
	})
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func (lgc *Logics) TencentCloudNewCredential(kit *rest.Kit, secretID, secretKey string) *tcCommon.Credential {
	return tcCommon.NewCredential(secretID, secretKey)
}

func (lgc *Logics) GetAwsVpc(kit *rest.Kit, secretID, secretKey string) ([]metadata.VpcInfo, error) {
	regions, err := lgc.GetAwsRegions(kit, secretID, secretKey)
	if err != nil {
		blog.ErrorJSON("getAwsVpc failed, because getAwsVpc failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	vpcs := make([]metadata.VpcInfo, 0)
	for _, region := range regions {
		sess, err := lgc.AwsNewSession(kit, region, secretID, secretKey)
		if err != nil {
			blog.ErrorJSON("getAwsVpc failed, awsNewSession failed, err: %v, rid: %s", err, kit.Rid)
			return nil, err
		}
		ec2Svc := ec2.New(sess)
		output, err := ec2Svc.DescribeVpcs(nil)
		for _, vpc := range output.Vpcs {
			vpcs = append(vpcs, metadata.VpcInfo{
				VpcName: *vpc.VpcId,
				VpcID:   *vpc.VpcId,
				Region:  region,
			})
		}
	}

	return vpcs, nil
}

func (lgc *Logics) GetTencentCloudVpc(kit *rest.Kit, secretID, secretKey string) ([]metadata.VpcInfo, error) {
	credential := lgc.TencentCloudNewCredential(kit, secretID, secretKey)
	regions, err := lgc.GetTencentCloudRegions(kit, secretID, secretKey)
	if err != nil {
		blog.ErrorJSON("getTencentCloudVpc failed, getTencentCloudRegions failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	vpcs := make([]metadata.VpcInfo, 0)
	for _, region := range regions {
		client, err := tcVpc.NewClient(credential, region, profile.NewClientProfile())
		if err != nil {
			blog.ErrorJSON("getTencentCloudVpc new client failed, err: %v, rid: %s", err, kit.Rid)
			return nil, nil
		}

		vpcReq := tcVpc.NewDescribeVpcsRequest()
		vpcResp, err := client.DescribeVpcs(vpcReq)
		for _, vpc := range vpcResp.Response.VpcSet {
			vpcs = append(vpcs, metadata.VpcInfo{
				VpcName: *vpc.VpcName,
				VpcID:   *vpc.VpcId,
				Region:  region,
			})
		}
	}

	return vpcs, nil
}

func (lgc *Logics) GetAwsInstance(kit *rest.Kit, secretID, secretKey string) {
	return
}

func (lgc *Logics) GetTencentCloudInstance(kit *rest.Kit, secretID, secretKey string) {
	return
}
