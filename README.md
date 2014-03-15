# Some useful AWS utilities

## Current utils

### Snapshot

Make snapshot backups of AWS instances. Created to have a simple binary distribution inspired by http://bit.ly/PDDaxI

#### Assumptions

  - there is no more than one EBS root and one additional EBS volume.
  - The Name 'tag' has some value in it.
  - instances to be bakced up have an env tag with the value "prod"

**NB**: these settings are untested and I migh thave spelled the
specifics incorrectly.

	example crontab entry to keep 7 daily snapshots
	0   3  *   *   *   /usr/local/bin/snapshot -count 7
	example crontab entry to keep 3 weekly snapshots
	0   5  */7  *  *   /usr/local/bin/snapshot -period weekly

**NB**: You will need to run the script with credentials that have something like
the following permissions -- either IAM creds or on an instance launched
into appropriate IAM role with permissions like the following:

	{
	  "Version": "2012-10-17",
	  "Statement": [
	    {
	      "Sid": "AllowEc2BackupSnapshotManagement",
	      "Effect": "Allow",
	      "Action": [
	        "ec2:CopySnapshot",
	        "ec2:CreateSnapshot",
	        "ec2:CreateTags",
	        "ec2:DeleteSnapshot",
	        "ec2:DescribeInstances",
	        "ec2:DescribeReservedInstances",
	        "ec2:DescribeSnapshotAttribute",
	        "ec2:DescribeSnapshots",
	        "ec2:DescribeTags",
	        "ec2:DescribeVolumes",
	        "ec2:ModifySnapshotAttribute"
	      ],
	      "Resource": [
	        "*"
	      ]
	    }
	  ]
	}

#### Known shortcomings

 1. Relies on cron for time deltas. Doesn't look at last backup time, just makes a new snapshot. This means running it repeatedly will delete older snapshots leaving you with all very recent snapshots. Ideally shouldn't create a new snapshot if hte last was less than *period*. This requires period be more than a string.

 2. Doesn't alert when there are problems. Ideally should send notifications to an SNS topic.... like http://bit.ly/PDDe0r
