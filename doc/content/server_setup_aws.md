# AWS-specific server setup instructions

These instructions are part of the instructions for
[Setting up `kikiserver`](/doc/server_setup.md).
Please make sure you have read that document first.

## Build `kikiserver-aws` and `kiki-setupstorage-aws`

To use Amazon Web Services fetch the `aws.kiki.io` repository and use the
`kikiserver-aws` and `kiki-setupstorage-aws` variants.

Fetch the repository and its dependencies:

```
local$ go get -d aws.kiki.io/cmd/...
```

Install the `kiki-setupstorage-aws` command:

```
local$ go install aws.kiki.io/cmd/kiki-setupstorage-aws
```

Build the `kikiserver-aws` binary:

```
local$ GOOS=linux GOARCH=amd64 go build aws.kiki.io/cmd/kikiserver-aws
```

## Install the AWS CLI

Ensure you have a working AWS environment set up before continuing and that you
are able to run basic commands using the
[CLI tool](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-welcome.html).

## Set up storage, role account, and instance profile

Use `kiki setupstorage-aws` to create an S3 bucket, an associated
role account, and instance profile for accessing the bucket and provisioning.
Note that the bucket name must be globally unique among all AWS users, so it is
prudent to include your domain name in the bucket name.
(We will use `example-com-kiki`.)

```
local$ kiki setupstorage-aws -domain=example.com example-com-kiki
```

It should produce output like this:

```
You should now deploy the kikiserver binary and run 'kiki setupserver'.
```

If the command fails, it may leave things in an incomplete state.
You can use the -clean flag to clean up any potential entities created:

```
local$ kiki setupstorage-aws -clean -role_name=kikistorage -domain=example.com example-com-kiki
```

**Notes**:

- The role has access to all S3 buckets by default. To restrict its access to
  only one bucket, follow [this guide](https://aws.amazon.com/blogs/security/how-to-restrict-amazon-s3-bucket-access-to-a-specific-iam-role/).
- The role name is also used as the name for the
  [instance profile](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2_instance-profiles.html)
  you should use to provision the instance.
- If you are running `kikiserver` on an EC2 instance, ensure that your
  security group allows inbound TCP traffic at least on port 443.

## Continue

You can now continue following the instructions in
[Setting up `kikiserver`](/doc/server_setup.md).
