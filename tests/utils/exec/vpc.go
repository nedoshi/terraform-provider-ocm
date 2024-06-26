package exec

import (
	"context"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type VPCArgs struct {
	Name                      string   `json:"name,omitempty"`
	AWSRegion                 string   `json:"aws_region,omitempty"`
	VPCCIDR                   string   `json:"vpc_cidr,omitempty"`
	MultiAZ                   bool     `json:"multi_az,omitempty"`
	AZIDs                     []string `json:"az_ids,omitempty"`
	HCP                       bool     `json:"hcp,omitempty"`
	AWSSharedCredentialsFiles []string `json:"aws_shared_credentials_files,omitempty"`
	DisableSubnetTagging      bool     `json:"disable_subnet_tagging,omitempty"`
}

type VPCOutput struct {
	ClusterPublicSubnets  []string `json:"cluster-public-subnet,omitempty"`
	VPCCIDR               string   `json:"vpc-cidr,omitempty"`
	ClusterPrivateSubnets []string `json:"cluster-private-subnet,omitempty"`
	AZs                   []string `json:"azs,omitempty"`
	NodePrivateSubnets    []string `json:"node-private-subnet,omitempty"`
	VPCID                 string   `json:"vpc_id,omitempty"`
}

type VPCService struct {
	CreationArgs *VPCArgs
	ManifestDir  string
	Context      context.Context
}

func (vpc *VPCService) Init(manifestDirs ...string) error {
	vpc.ManifestDir = CON.AWSVPCDir
	if len(manifestDirs) != 0 {
		vpc.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	vpc.Context = ctx
	err := runTerraformInit(ctx, vpc.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (vpc *VPCService) Apply(createArgs *VPCArgs, recordtfvars bool, extraArgs ...string) error {
	vpc.CreationArgs = createArgs
	args, tfvars := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApply(vpc.Context, vpc.ManifestDir, args...)
	if err != nil {
		return err
	}
	if recordtfvars {
		recordTFvarsFile(vpc.ManifestDir, tfvars)
	}

	return nil
}

func (vpc *VPCService) Output() (*VPCOutput, error) {
	vpcDir := CON.AWSVPCDir
	if vpc.ManifestDir != "" {
		vpcDir = vpc.ManifestDir
	}
	out, err := runTerraformOutput(context.TODO(), vpcDir)
	if err != nil {
		return nil, err
	}
	vpcOutput := &VPCOutput{
		VPCCIDR:               h.DigString(out["vpc-cidr"], "value"),
		ClusterPrivateSubnets: h.DigArrayToString(out["cluster-private-subnet"], "value"),
		ClusterPublicSubnets:  h.DigArrayToString(out["cluster-public-subnet"], "value"),
		NodePrivateSubnets:    h.DigArrayToString(out["node-private-subnet"], "value"),
		AZs:                   h.DigArrayToString(out["azs"], "value"),
		VPCID:                 h.DigString(out["vpc-id"], "value"),
	}

	return vpcOutput, err
}

func (vpc *VPCService) Destroy(createArgs ...*VPCArgs) error {
	if vpc.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := vpc.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	args, _ := combineStructArgs(destroyArgs)
	_, err := runTerraformDestroy(vpc.Context, vpc.ManifestDir, args...)

	return err
}

func NewVPCService(manifestDir ...string) *VPCService {
	vpc := &VPCService{}
	vpc.Init(manifestDir...)
	return vpc
}
