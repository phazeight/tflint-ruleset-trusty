// rules/trusty_key_attributes_test.go
package rules_test

import (
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

// These cases pin the ordering contract for the EKS family: every key attribute
// first (contiguously), then the remainder A-Z.
//
// Note for atom authors: scaffold-atom.py only special-cases `name`, so for
// resources whose key attributes are named otherwise (aws_eks_addon,
// aws_eks_access_entry, aws_eks_node_group) the generated main.tf needs manual
// reordering before it satisfies this rule.
func TestKeyAttributesRule_EksFamily(t *testing.T) {
	rule := rules.NewKeyAttributesRule()

	cases := []struct {
		name      string
		files     map[string]string
		wantCount int
	}{
		{
			name: "eks-cluster — name first, remainder A-Z",
			files: map[string]string{
				"modules/aws/atoms/eks-cluster/main.tf": `
resource "aws_eks_cluster" "this" {
  name                          = var.name
  bootstrap_self_managed_addons = var.bootstrap_self_managed_addons
  deletion_protection           = var.deletion_protection
  region                        = var.region
  role_arn                      = var.role_arn
  tags                          = local.tags
  version                       = var.kubernetes_version
}`,
			},
			wantCount: 0,
		},
		{
			name: "eks-addon — both key attrs first, then A-Z",
			files: map[string]string{
				"modules/aws/atoms/eks-addon/main.tf": `
resource "aws_eks_addon" "this" {
  cluster_name  = var.cluster_name
  addon_name    = var.addon_name
  addon_version = var.addon_version
  preserve      = var.preserve
  tags          = local.tags
}`,
			},
			wantCount: 0,
		},
		{
			name: "eks-addon — non-key attr wedged between the key attrs",
			files: map[string]string{
				"modules/aws/atoms/eks-addon/main.tf": `
resource "aws_eks_addon" "this" {
  addon_name    = var.addon_name
  addon_version = var.addon_version
  cluster_name  = var.cluster_name
  tags          = local.tags
}`,
			},
			wantCount: 1,
		},
		{
			name: "eks-access-entry — both key attrs first, then A-Z",
			files: map[string]string{
				"modules/aws/atoms/eks-access-entry/main.tf": `
resource "aws_eks_access_entry" "this" {
  cluster_name      = var.cluster_name
  principal_arn     = var.principal_arn
  kubernetes_groups = var.kubernetes_groups
  region            = var.region
  tags              = local.tags
}`,
			},
			wantCount: 0,
		},
		{
			name: "eks-node-group — both key attrs first, then A-Z",
			files: map[string]string{
				"modules/aws/atoms/eks-node-group/main.tf": `
resource "aws_eks_node_group" "this" {
  cluster_name    = var.cluster_name
  node_group_name = var.node_group_name
  ami_type        = var.ami_type
  capacity_type   = var.capacity_type
  disk_size       = var.disk_size
  instance_types  = var.instance_types
  node_role_arn   = var.node_role_arn
  tags            = local.tags
}`,
			},
			wantCount: 0,
		},
		{
			name: "violation — non-key attribute before the key attribute",
			files: map[string]string{
				"modules/aws/atoms/eks-cluster/main.tf": `
resource "aws_eks_cluster" "this" {
  role_arn = var.role_arn
  name     = var.name
}`,
			},
			wantCount: 1,
		},
		{
			name: "violation — remainder not sorted A-Z",
			files: map[string]string{
				"modules/aws/atoms/eks-cluster/main.tf": `
resource "aws_eks_cluster" "this" {
  name     = var.name
  tags     = local.tags
  role_arn = var.role_arn
}`,
			},
			wantCount: 1,
		},
		{
			name: "unknown resource type is skipped",
			files: map[string]string{
				"modules/aws/atoms/whatever/main.tf": `
resource "aws_totally_unknown_thing" "this" {
  zebra = 1
  apple = 2
}`,
			},
			wantCount: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, tc.files)
			if err := rule.Check(runner); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if len(runner.Issues) != tc.wantCount {
				t.Errorf("got %d issue(s), want %d", len(runner.Issues), tc.wantCount)
				for _, iss := range runner.Issues {
					t.Logf("  - %s", iss.Message)
				}
			}
		})
	}
}

// These cases pin the ordering contract for the EFS family.
//
// Four of the five take file_system_id as their key attribute and aws_efs_file_system
// takes creation_token, so none of them is the `name` case scaffold-atom.py
// special-cases — every generated main.tf here needs manual reordering before it
// satisfies this rule, the same trap the EKS atoms hit.
func TestKeyAttributesRule_EfsFamily(t *testing.T) {
	rule := rules.NewKeyAttributesRule()

	cases := []struct {
		name      string
		files     map[string]string
		wantCount int
	}{
		{
			name: "efs-file-system — creation_token first, remainder A-Z",
			files: map[string]string{
				"modules/aws/atoms/efs-file-system/main.tf": `
resource "aws_efs_file_system" "this" {
  creation_token                  = var.creation_token
  availability_zone_name          = var.availability_zone_name
  encrypted                       = var.encrypted
  kms_key_id                      = var.kms_key_id
  performance_mode                = var.performance_mode
  provisioned_throughput_in_mibps = var.provisioned_throughput_in_mibps
  region                          = var.region
  tags                            = local.tags
  throughput_mode                 = var.throughput_mode
}`,
			},
			wantCount: 0,
		},
		{
			name: "efs-mount-target — both key attrs first, then A-Z",
			files: map[string]string{
				"modules/aws/atoms/efs-mount-target/main.tf": `
resource "aws_efs_mount_target" "this" {
  file_system_id  = var.file_system_id
  subnet_id       = var.subnet_id
  ip_address      = var.ip_address
  region          = var.region
  security_groups = var.security_groups
}`,
			},
			wantCount: 0,
		},
		{
			name: "efs-mount-target — non-key attr wedged between the key attrs",
			files: map[string]string{
				"modules/aws/atoms/efs-mount-target/main.tf": `
resource "aws_efs_mount_target" "this" {
  file_system_id  = var.file_system_id
  ip_address      = var.ip_address
  subnet_id       = var.subnet_id
  security_groups = var.security_groups
}`,
			},
			wantCount: 1,
		},
		{
			name: "efs-access-point — file_system_id first, remainder A-Z",
			files: map[string]string{
				"modules/aws/atoms/efs-access-point/main.tf": `
resource "aws_efs_access_point" "this" {
  file_system_id = var.file_system_id
  region         = var.region
  tags           = local.tags
}`,
			},
			wantCount: 0,
		},
		{
			name: "efs-file-system-policy — file_system_id before policy",
			files: map[string]string{
				"modules/aws/atoms/efs-file-system-policy/main.tf": `
resource "aws_efs_file_system_policy" "this" {
  file_system_id                     = var.file_system_id
  bypass_policy_lockout_safety_check = var.bypass_policy_lockout_safety_check
  policy                             = var.policy
  region                             = var.region
}`,
			},
			wantCount: 0,
		},
		{
			name: "efs-file-system-policy — policy before the key attr",
			files: map[string]string{
				"modules/aws/atoms/efs-file-system-policy/main.tf": `
resource "aws_efs_file_system_policy" "this" {
  policy         = var.policy
  file_system_id = var.file_system_id
  region         = var.region
}`,
			},
			wantCount: 1,
		},
		{
			name: "efs-backup-policy — file_system_id first",
			files: map[string]string{
				"modules/aws/atoms/efs-backup-policy/main.tf": `
resource "aws_efs_backup_policy" "this" {
  file_system_id = var.file_system_id
  region         = var.region
}`,
			},
			wantCount: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, tc.files)
			if err := rule.Check(runner); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if len(runner.Issues) != tc.wantCount {
				t.Errorf("got %d issue(s), want %d", len(runner.Issues), tc.wantCount)
				for _, iss := range runner.Issues {
					t.Logf("  - %s", iss.Message)
				}
			}
		})
	}
}
