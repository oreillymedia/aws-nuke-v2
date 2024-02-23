//go:generate ../mocks/generate_mocks.sh cloudformation cloudformationiface
package resources

import _ "github.com/golang/mock/mockgen"

// Note: empty on purpose, this file exist purely to generate mocks for the CloudFormation service
