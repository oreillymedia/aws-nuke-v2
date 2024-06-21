package resources

import (
	"context"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/aws/aws-sdk-go/service/backup"
)

type BackupVaultAccessPolicy struct {
	svc             *backup.Backup
	backupVaultName string
}

const AWSBackupVaultAccessPolicyResource = "AWSBackupVaultAccessPolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:   AWSBackupVaultAccessPolicyResource,
		Scope:  nuke.Account,
		Lister: &AWSBackupVaultAccessPolicyLister{},
	})
}

type AWSBackupVaultAccessPolicyLister struct{}

func (l *AWSBackupVaultAccessPolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := backup.New(opts.Session)
	maxVaultsLen := int64(100)
	params := &backup.ListBackupVaultsInput{
		MaxResults: &maxVaultsLen, // aws default limit on number of backup vaults per account
	}
	resp, err := svc.ListBackupVaults(params)
	if err != nil {
		return nil, err
	}

	// Iterate over backup vaults and add vault policies that exist.
	resources := make([]resource.Resource, 0)
	for _, out := range resp.BackupVaultList {
		// Check if the Backup Vault has an Access Policy set
		resp, err := svc.GetBackupVaultAccessPolicy(&backup.GetBackupVaultAccessPolicyInput{
			BackupVaultName: out.BackupVaultName,
		})

		// Non-existent Access Policies can come from ResourceNotFoundException or
		// being nil.
		if err != nil {
			switch err.(type) {
			case *backup.ResourceNotFoundException:
				// Non-existent is OK and we skip over them
				continue
			default:
				return nil, err
			}
		}

		// Only delete policies that exist.
		if resp.Policy != nil {
			resources = append(resources, &BackupVaultAccessPolicy{
				svc:             svc,
				backupVaultName: *out.BackupVaultName,
			})
		}
	}

	return resources, nil
}

func (b *BackupVaultAccessPolicy) Remove() error {
	// Set a policy that allows deletion before removal.
	//
	// This is required to delete the policy for the automagically created vaults
	// such as "aws/efs/automatic-backup-vault" from EFS automatic backups
	// which by default Deny policy deletion via backup::DeleteBackupVaultAccessPolicy
	//
	// Example "aws/efs/automatic-backup-vault" default policy:
	//
	// {
	//     "Version": "2012-10-17",
	//     "Statement": [
	//         {
	//             "Effect": "Deny",
	//             "Principal": {
	//                 "AWS": "*"
	//             },
	//             "Action": [
	//                 "backup:DeleteBackupVault",
	//                 "backup:DeleteBackupVaultAccessPolicy",
	//                 "backup:DeleteRecoveryPoint",
	//                 "backup:StartCopyJob",
	//                 "backup:StartRestoreJob",
	//                 "backup:UpdateRecoveryPointLifecycle"
	//             ],
	//             "Resource": "*"
	//         }
	//     ]
	// }
	//
	// Update the default policy to remove the Deny on Delete* actions
	// and then delete the policy.
	//
	// Why not putting a policy that allows `backup:DeleteBackupVaultAccessPolicy` in the first place?
	// Because that throws an error:
	// ' The specified policy cannot be added to the vault due to cross-account sharing restrictions.
	// Amend the policy or the vault's settings, then retry request'
	//
	allowDeletionPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Deny",
				"Principal": {
					"AWS": "*"
				},
				"Action": [
					"backup:StartCopyJob",
					"backup:StartRestoreJob",
					"backup:UpdateRecoveryPointLifecycle"
				],
				"Resource": "*"
			}
		]
	}`
	// Ignore error from if we can't put permissive backup vault policy in for some reason, that's OK.
	_, _ = b.svc.PutBackupVaultAccessPolicy(&backup.PutBackupVaultAccessPolicyInput{
		BackupVaultName: &b.backupVaultName,
		Policy:          &allowDeletionPolicy,
	})
	// In the end, this is the only call we actually really care about for err.
	_, err := b.svc.DeleteBackupVaultAccessPolicy(&backup.DeleteBackupVaultAccessPolicyInput{
		BackupVaultName: &b.backupVaultName,
	})
	return err
}

func (b *BackupVaultAccessPolicy) String() string {
	return b.backupVaultName
}
