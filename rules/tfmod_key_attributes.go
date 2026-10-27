package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// defaultKeyAttributes maps resource type to the attribute names that identify the
// resource and must appear first in the resource block, before all other attributes.
var defaultKeyAttributes = map[string][]string{
	"aws_acm_certificate":                   {"domain_name"},
	"aws_cloudwatch_log_group":              {"name"},
	"aws_cognito_user_pool":                 {"name"},
	"aws_db_instance":                       {"identifier"},
	"aws_db_subnet_group":                   {"name"},
	"aws_dynamodb_table":                    {"name"},
	"aws_ecr_repository":                    {"name"},
	"aws_ecs_cluster":                       {"name"},
	"aws_ecs_task_definition":               {"family"},
	"aws_ecs_service":                       {"name"},
	"aws_efs_access_point":                  {"file_system_id"},
	"aws_efs_backup_policy":                 {"file_system_id"},
	"aws_efs_file_system":                   {"creation_token"},
	"aws_efs_file_system_policy":            {"file_system_id"},
	"aws_efs_mount_target":                  {"file_system_id", "subnet_id"},
	"aws_eip":                               {"domain"},
	"aws_eks_access_entry":                  {"cluster_name", "principal_arn"},
	"aws_eks_addon":                         {"cluster_name", "addon_name"},
	"aws_eks_cluster":                       {"name"},
	"aws_eks_node_group":                    {"cluster_name", "node_group_name"},
	"aws_elasticache_cluster":               {"cluster_id"},
	"aws_flow_log":                          {"vpc_id"},
	"aws_iam_group":                         {"name"},
	"aws_iam_instance_profile":              {"name"},
	"aws_iam_openid_connect_provider":       {"url"},
	"aws_iam_policy":                        {"name"},
	"aws_iam_policy_attachment":             {"name"},
	"aws_iam_role":                          {"name"},
	"aws_iam_role_policy":                   {"name"},
	"aws_iam_role_policy_attachment":        {"role"},
	"aws_iam_user":                          {"name"},
	"aws_iam_user_policy_attachment":        {"user"},
	"aws_internet_gateway":                  {"vpc_id"},
	"aws_kinesis_stream":                    {"name"},
	"aws_kms_alias":                         {"name"},
	"aws_kms_key":                           {"description"},
	"aws_lambda_function":                   {"function_name"},
	"aws_lambda_layer_version":              {"layer_name"},
	"aws_launch_template":                   {"name"},
	"aws_lb":                                {"name"},
	"aws_lb_listener":                       {"load_balancer_arn"},
	"aws_lb_target_group":                   {"name"},
	"aws_nat_gateway":                       {"subnet_id"},
	"aws_rds_cluster":                       {"cluster_identifier"},
	"aws_rds_cluster_instance":              {"identifier"},
	"aws_rds_cluster_parameter_group":       {"name"},
	"aws_route":                             {"route_table_id"},
	"aws_route53_record":                    {"name"},
	"aws_route53_zone":                      {"name"},
	"aws_route_table":                       {"vpc_id"},
	"aws_s3_bucket":                         {"bucket"},
	"aws_s3_bucket_acl":                     {"bucket"},
	"aws_s3_bucket_cors_configuration":      {"bucket"},
	"aws_s3_bucket_lifecycle_configuration": {"bucket"},
	"aws_s3_bucket_notification":            {"bucket"},
	"aws_s3_bucket_ownership_controls":      {"bucket"},
	"aws_s3_bucket_policy":                  {"bucket"},
	"aws_s3_bucket_public_access_block":     {"bucket"},
	"aws_s3_bucket_server_side_encryption_configuration": {"bucket"},
	"aws_s3_bucket_versioning":                           {"bucket"},
	"aws_s3_object":                                      {"bucket", "key"},
	"aws_secretsmanager_secret":                          {"name"},
	"aws_security_group":                                 {"name"},
	"aws_sns_topic":                                      {"name"},
	"aws_sqs_queue":                                      {"name"},
	"aws_ssm_parameter":                                  {"name"},
	"aws_subnet":                                         {"vpc_id"},
	"aws_vpc":                                            {"cidr_block"},
	"aws_wafv2_web_acl":                                  {"name"},
}

// KeyAttributesRule enforces that identifying attributes in resource blocks appear
// before all other attributes, and that remaining attributes are sorted A-Z.
type KeyAttributesRule struct {
	tflint.DefaultRule
}

func NewKeyAttributesRule() *KeyAttributesRule {
	return &KeyAttributesRule{}
}

func (r *KeyAttributesRule) Name() string              { return "trusty_key_attributes" }
func (r *KeyAttributesRule) Enabled() bool             { return true }
func (r *KeyAttributesRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *KeyAttributesRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_key_attributes.md"
}

type attrPos struct {
	name string
	line int
	rng  hcl.Range
}

func (r *KeyAttributesRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}

		for _, block := range body.Blocks {
			if block.Type != "resource" || len(block.Labels) < 2 {
				continue
			}
			resourceType := block.Labels[0]
			resourceName := block.Labels[1]

			keyAttrs, known := defaultKeyAttributes[resourceType]
			if !known || len(keyAttrs) == 0 {
				continue
			}
			keySet := make(map[string]bool, len(keyAttrs))
			for _, k := range keyAttrs {
				keySet[k] = true
			}

			// Collect all attributes with source positions
			var attrs []attrPos
			for name, attr := range block.Body.Attributes {
				attrs = append(attrs, attrPos{
					name: name,
					line: attr.NameRange.Start.Line,
					rng:  attr.NameRange,
				})
			}
			sort.Slice(attrs, func(i, j int) bool {
				return attrs[i].line < attrs[j].line
			})

			if len(attrs) == 0 {
				continue
			}

			// Find the last line at which any key attribute appears
			lastKeyLine := -1
			for _, a := range attrs {
				if keySet[a.name] && a.line > lastKeyLine {
					lastKeyLine = a.line
				}
			}
			if lastKeyLine == -1 {
				// None of the configured key attributes are present — skip
				continue
			}

			// Any non-key attribute that appears before the last key attribute is a violation
			for _, a := range attrs {
				if keySet[a.name] {
					continue
				}
				if a.line < lastKeyLine {
					if err := runner.EmitIssue(r,
						fmt.Sprintf(
							`in resource "%s" "%s": attribute "%s" must appear after key attribute(s) (%s)`,
							resourceType, resourceName, a.name,
							strings.Join(keyAttrs, ", "),
						),
						a.rng,
					); err != nil {
						return err
					}
				}
			}

			// Non-key attributes after the key section must be sorted A-Z
			var nonKeyAfter []attrPos
			for _, a := range attrs {
				if !keySet[a.name] && a.line > lastKeyLine {
					nonKeyAfter = append(nonKeyAfter, a)
				}
			}
			for i := 1; i < len(nonKeyAfter); i++ {
				if nonKeyAfter[i].name < nonKeyAfter[i-1].name {
					if err := runner.EmitIssue(r,
						fmt.Sprintf(
							`in resource "%s" "%s": attribute "%s" must be sorted A-Z (appears after "%s")`,
							resourceType, resourceName, nonKeyAfter[i].name, nonKeyAfter[i-1].name,
						),
						nonKeyAfter[i].rng,
					); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
