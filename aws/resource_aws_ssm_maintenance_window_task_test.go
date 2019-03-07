package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSSMMaintenanceWindowTask_basic(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.target"

	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTaskBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTask_updateForcesNewResource(t *testing.T) {
	var before, after ssm.MaintenanceWindowTask
	name := acctest.RandString(10)
	resourceName := "aws_ssm_maintenance_window_task.target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTaskBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &before),
				),
			},
			{
				Config: testAccAWSSSMMaintenanceWindowTaskBasicConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", "TestMaintenanceWindowTask"),
					resource.TestCheckResourceAttr(resourceName, "description", "This resource is for test purpose only"),
					testAccCheckAwsSsmWindowsTaskRecreated(t, &before, &after),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTask_TaskInvocationAutomationParameters(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.target"

	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTaskAutomationConfig(name, "$DEFAULT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.automation_parameters.0.document_version", "$DEFAULT"),
				),
			},
			{
				Config: testAccAWSSSMMaintenanceWindowTaskAutomationConfigUpdate(name, "$LATEST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.automation_parameters.0.document_version", "$LATEST"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTask_TaskInvocationLambdaParameters(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.target"
	rString := acctest.RandString(8)
	rInt := acctest.RandInt()

	funcName := fmt.Sprintf("tf_acc_lambda_func_tags_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_tags_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_tags_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_tags_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTaskLambdaConfig(funcName, policyName, roleName, sgName, rString, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTask_TaskInvocationRunCommandParameters(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.target"
	serviceRoleResourceName := "aws_iam_role.ssm_role"
	s3BucketResourceName := "aws_s3_bucket.foo"

	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTaskRunCommandConfig(name, "test comment", 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttrPair(resourceName, "service_role_arn", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "task_invocation_parameters.0.run_command_parameters.0.service_role_arn", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.comment", "test comment"),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.timeout_seconds", "30"),
				),
			},
			{
				Config: testAccAWSSSMMaintenanceWindowTaskRunCommandConfigUpdate(name, "test comment update", 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.comment", "test comment update"),
					resource.TestCheckResourceAttr(resourceName, "task_invocation_parameters.0.run_command_parameters.0.timeout_seconds", "60"),
					resource.TestCheckResourceAttrPair(resourceName, "task_invocation_parameters.0.run_command_parameters.0.output_s3_bucket", s3BucketResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTask_TaskInvocationStepFunctionParameters(t *testing.T) {
	var task ssm.MaintenanceWindowTask
	resourceName := "aws_ssm_maintenance_window_task.target"
	rString := acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTaskStepFunctionConfig(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTaskExists(resourceName, &task),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsSsmWindowsTaskRecreated(t *testing.T,
	before, after *ssm.MaintenanceWindowTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.WindowTaskId == after.WindowTaskId {
			t.Fatalf("Expected change of Windows Task IDs, but both were %v", before.WindowTaskId)
		}
		return nil
	}
}

func testAccCheckAWSSSMMaintenanceWindowTaskExists(n string, task *ssm.MaintenanceWindowTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Maintenance Window Task Window ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		resp, err := conn.DescribeMaintenanceWindowTasks(&ssm.DescribeMaintenanceWindowTasksInput{
			WindowId: aws.String(rs.Primary.Attributes["window_id"]),
		})
		if err != nil {
			return err
		}

		for _, i := range resp.Tasks {
			if *i.WindowTaskId == rs.Primary.ID {
				*task = *i
				return nil
			}
		}

		return fmt.Errorf("No AWS SSM Maintenance window task found")
	}
}

func testAccCheckAWSSSMMaintenanceWindowTaskDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_maintenance_window_target" {
			continue
		}

		out, err := conn.DescribeMaintenanceWindowTasks(&ssm.DescribeMaintenanceWindowTasksInput{
			WindowId: aws.String(rs.Primary.Attributes["window_id"]),
		})

		if err != nil {
			// Verify the error is what we want
			if ae, ok := err.(awserr.Error); ok && ae.Code() == "DoesNotExistException" {
				continue
			}
			return err
		}

		if len(out.Tasks) > 0 {
			return fmt.Errorf("Expected AWS SSM Maintenance Task to be gone, but was still found")
		}

		return nil
	}

	return nil
}

func testAccAWSSSMMaintenanceWindowTaskImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["window_id"], rs.Primary.ID), nil
	}
}

func testAccAWSSSMMaintenanceWindowTaskBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "foo" {
  name = "maintenance-window-%s"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff = 1
}

resource "aws_ssm_maintenance_window_task" "target" {
  window_id = "${aws_ssm_maintenance_window.foo.id}"
  task_type = "RUN_COMMAND"
  task_arn = "AWS-RunShellScript"
  priority = 1
  service_role_arn = "${aws_iam_role.ssm_role.arn}"
  max_concurrency = "2"
  max_errors = "1"
  targets {
    key = "InstanceIds"
    values = ["${aws_instance.foo.id}"]
  }
  task_parameters {
    name = "commands"
    values = ["pwd"]
  }
}

resource "aws_instance" "foo" {
  ami = "ami-4fccb37f"

  instance_type = "m1.small"
}

resource "aws_iam_role" "ssm_role" {
  name = "ssm-role-%s"

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
                "Service": "events.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy" "bar" {
  name = "ssm_role_policy_%s"
  role = "${aws_iam_role.ssm_role.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ssm:*",
    "Resource": "*"
  }
}
EOF
}

`, rName, rName, rName)
}

func testAccAWSSSMMaintenanceWindowTaskBasicConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "foo" {
  name = "maintenance-window-%s"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff = 1
}

resource "aws_ssm_maintenance_window_task" "target" {
  window_id = "${aws_ssm_maintenance_window.foo.id}"
  task_type = "RUN_COMMAND"
  task_arn = "AWS-RunShellScript"
  priority = 1
  name = "TestMaintenanceWindowTask"
  description = "This resource is for test purpose only"
  service_role_arn = "${aws_iam_role.ssm_role.arn}"
  max_concurrency = "2"
  max_errors = "1"
  targets {
    key = "InstanceIds"
    values = ["${aws_instance.foo.id}"]
  }
  task_parameters {
    name = "commands"
    values = ["date"]
  }
}

resource "aws_instance" "foo" {
  ami = "ami-4fccb37f"

  instance_type = "m1.small"
}

resource "aws_iam_role" "ssm_role" {
  name = "ssm-role-%s"

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
                "Service": "events.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy" "bar" {
  name = "ssm_role_policy_%s"
  role = "${aws_iam_role.ssm_role.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ssm:*",
    "Resource": "*"
  }
}
EOF
}

`, rName, rName, rName)
}

func testAccAWSSSMMaintenanceWindowTaskAutomationConfig(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "foo" {
  name = "maintenance-window-%[1]s"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff = 1
}

resource "aws_ssm_maintenance_window_task" "target" {
  window_id = "${aws_ssm_maintenance_window.foo.id}"
  task_type = "AUTOMATION"
  task_arn = "AWS-CreateImage"
  priority = 1
  service_role_arn = "${aws_iam_role.ssm_role.arn}"
  max_concurrency = "2"
  max_errors = "1"
  targets {
    key = "InstanceIds"
    values = ["${aws_instance.foo.id}"]
  }
  task_invocation_parameters {
		automation_parameters {
			document_version = "%[2]s"
			parameters {
				name = "InstanceId"
				values = ["${aws_instance.foo.id}"]
			}
			parameters {
				name = "NoReboot"
				values = ["false"]
			}
		}
  }
}

resource "aws_instance" "foo" {
  ami = "ami-4fccb37f"

  instance_type = "m1.small"
}

resource "aws_iam_role" "ssm_role" {
  name = "ssm-role-%[1]s"

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
                "Service": "events.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy" "bar" {
  name = "ssm_role_policy_%[1]s"
  role = "${aws_iam_role.ssm_role.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ssm:*",
    "Resource": "*"
  }
}
EOF
}

`, rName, version)
}

func testAccAWSSSMMaintenanceWindowTaskAutomationConfigUpdate(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "foo" {
		bucket = "tf-s3-%[1]s"
		acl = "private"
		force_destroy = true
}

resource "aws_ssm_maintenance_window" "foo" {
  name = "maintenance-window-%[1]s"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff = 1
}

resource "aws_ssm_maintenance_window_task" "target" {
  window_id = "${aws_ssm_maintenance_window.foo.id}"
  task_type = "AUTOMATION"
  task_arn = "AWS-CreateImage"
  priority = 1
  service_role_arn = "${aws_iam_role.ssm_role.arn}"
  max_concurrency = "2"
  max_errors = "1"
  targets {
    key = "InstanceIds"
    values = ["${aws_instance.foo.id}"]
  }
  task_invocation_parameters {
		automation_parameters {
			document_version = "%[2]s"
			parameters {
				name = "InstanceId"
				values = ["${aws_instance.foo.id}"]
			}
			parameters {
				name = "NoReboot"
				values = ["false"]
			}
		}
  }
}

resource "aws_instance" "foo" {
  ami = "ami-4fccb37f"

  instance_type = "m1.small"
}

resource "aws_iam_role" "ssm_role" {
  name = "ssm-role-%[1]s"

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
                "Service": "events.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy" "bar" {
  name = "ssm_role_policy_%[1]s"
  role = "${aws_iam_role.ssm_role.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ssm:*",
    "Resource": "*"
  }
}
EOF
}

`, rName, version)
}

func testAccAWSSSMMaintenanceWindowTaskLambdaConfig(funcName, policyName, roleName, sgName, rName string, rInt int) string {
	return fmt.Sprintf(testAccAWSLambdaConfigBasic(funcName, policyName, roleName, sgName)+`
resource "aws_ssm_maintenance_window" "foo" {
  name = "maintenance-window-%[1]s"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff = 1
}

resource "aws_ssm_maintenance_window_task" "target" {
  window_id = "${aws_ssm_maintenance_window.foo.id}"
  task_type = "LAMBDA"
  task_arn = "${aws_lambda_function.lambda_function_test.arn}"
  priority = 1
  service_role_arn = "${aws_iam_role.ssm_role.arn}"
  max_concurrency = "2"
  max_errors = "1"
  targets {
    key = "InstanceIds"
    values = ["${aws_instance.foo.id}"]
  }
  task_invocation_parameters {
		lambda_parameters {
			client_context = "${base64encode(data.template_file.client_context.rendered)}"
			payload = "${data.template_file.payload.rendered}"
		}
  }
}

resource "aws_instance" "foo" {
  ami = "ami-4fccb37f"

  instance_type = "m1.small"
}

resource "aws_iam_role" "ssm_role" {
  name = "ssm-role-%[1]s"

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
                "Service": "events.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy" "bar" {
  name = "ssm_role_policy_%[1]s"
  role = "${aws_iam_role.ssm_role.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ssm:*",
    "Resource": "*"
  }
}
EOF
}

data "template_file" "client_context" {
  template = <<EOF
{
	"key1": "value1",
	"key2": "value2",
	"key3": "value3"
}
EOF
}

data "template_file" "payload" {
  template = <<EOF
{
	"number": %[2]d
}
EOF
}

`, rName, rInt)
}

func testAccAWSSSMMaintenanceWindowTaskRunCommandConfig(rName, comment string, timeoutSeconds int) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "foo" {
  name = "maintenance-window-%[1]s"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff = 1
}

resource "aws_ssm_maintenance_window_task" "target" {
  window_id = "${aws_ssm_maintenance_window.foo.id}"
  task_type = "RUN_COMMAND"
  task_arn = "AWS-RunShellScript"
  priority = 1
  service_role_arn = "${aws_iam_role.ssm_role.arn}"
  max_concurrency = "2"
  max_errors = "1"
  targets {
    key = "InstanceIds"
    values = ["${aws_instance.foo.id}"]
  }
  task_invocation_parameters {
		run_command_parameters {
			comment 						= "%[2]s"
			document_hash 			= "${sha256("COMMAND")}"
			document_hash_type 	= "Sha256"
			service_role_arn 		= "${aws_iam_role.ssm_role.arn}"
			timeout_seconds			= %[3]d
			parameters {
				name = "commands"
				values = ["date"]
			}
		}
  }
}

resource "aws_instance" "foo" {
  ami = "ami-4fccb37f"

  instance_type = "m1.small"
}

resource "aws_iam_role" "ssm_role" {
  name = "ssm-role-%[1]s"

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
                "Service": "events.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy" "bar" {
  name = "ssm_role_policy_%[1]s"
  role = "${aws_iam_role.ssm_role.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ssm:*",
    "Resource": "*"
  }
}
EOF
}

`, rName, comment, timeoutSeconds)
}

func testAccAWSSSMMaintenanceWindowTaskRunCommandConfigUpdate(rName, comment string, timeoutSeconds int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "foo" {
		bucket = "tf-s3-%[1]s"
		acl = "private"
		force_destroy = true
}

resource "aws_ssm_maintenance_window" "foo" {
  name = "maintenance-window-%[1]s"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff = 1
}

resource "aws_ssm_maintenance_window_task" "target" {
  window_id = "${aws_ssm_maintenance_window.foo.id}"
  task_type = "RUN_COMMAND"
  task_arn = "AWS-RunShellScript"
  priority = 1
  service_role_arn = "${aws_iam_role.ssm_role.arn}"
  max_concurrency = "2"
  max_errors = "1"
  targets {
    key = "InstanceIds"
    values = ["${aws_instance.foo.id}"]
  }
  task_invocation_parameters {
		run_command_parameters {
			comment 							= "%[2]s"
			document_hash 				= "${sha256("COMMAND")}"
			document_hash_type 		= "Sha256"
			service_role_arn 			= "${aws_iam_role.ssm_role.arn}"
			timeout_seconds				= %[3]d
			output_s3_bucket			= "${aws_s3_bucket.foo.id}"
			output_s3_key_prefix	= "foo"
			parameters {
				name = "commands"
				values = ["date"]
			}
		}
  }
}

resource "aws_instance" "foo" {
  ami = "ami-4fccb37f"

  instance_type = "m1.small"
}

resource "aws_iam_role" "ssm_role" {
  name = "ssm-role-%[1]s"

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
                "Service": "events.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy" "bar" {
  name = "ssm_role_policy_%[1]s"
  role = "${aws_iam_role.ssm_role.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ssm:*",
    "Resource": "*"
  }
}
EOF
}

`, rName, comment, timeoutSeconds)
}

func testAccAWSSSMMaintenanceWindowTaskStepFunctionConfig(rName string) string {
	return fmt.Sprintf(testAccAWSSfnActivityBasicConfig(rName)+`
resource "aws_ssm_maintenance_window" "foo" {
  name = "maintenance-window-%[1]s"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff = 1
}

resource "aws_ssm_maintenance_window_task" "target" {
  window_id = "${aws_ssm_maintenance_window.foo.id}"
  task_type = "STEP_FUNCTIONS"
  task_arn = "${aws_sfn_activity.foo.id}"
  priority = 1
  service_role_arn = "${aws_iam_role.ssm_role.arn}"
  max_concurrency = "2"
  max_errors = "1"
  targets {
    key = "InstanceIds"
    values = ["${aws_instance.foo.id}"]
  }
  task_invocation_parameters {
		step_functions_parameters {
			input = "${data.template_file.input.rendered}"
			name = "tf-step-function-%[1]s"
		}
  }
}

resource "aws_instance" "foo" {
  ami = "ami-4fccb37f"

  instance_type = "m1.small"
}

resource "aws_iam_role" "ssm_role" {
  name = "ssm-role-%[1]s"

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
                "Service": "events.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy" "bar" {
  name = "ssm_role_policy_%[1]s"
  role = "${aws_iam_role.ssm_role.name}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "ssm:*",
    "Resource": "*"
  }
}
EOF
}

data "template_file" "input" {
  template = <<EOF
{
	"key1": "value1",
	"key2": "value2",
	"key3": "value3"
}
EOF
}

`, rName)
}
