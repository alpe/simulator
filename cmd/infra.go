package cmd

import (
	"github.com/controlplaneio/simulator-standalone/pkg/simulator"
	"github.com/controlplaneio/simulator-standalone/pkg/ssh"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func newCreateCommand(logger *zap.SugaredLogger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   `create`,
		Short: "Runs terraform to create the required infrastructure for scenarios",
		RunE: func(cmd *cobra.Command, args []string) error {
			bucketName := viper.GetString("state-bucket")
			if bucketName == "" {
				logger.Warnf("Simulator has not been initialised with an S3 bucket")
				logger.Warn("Please run simulator init")
				return nil
			}

			scenariosDir := viper.GetString("scenarios-dir")
			attackTag := viper.GetString("attack-container-tag")
			tfDir := viper.GetString("tf-dir")
			tfVarsDir := viper.GetString("tf-vars-dir")

			logger.Infof("Created s3 bucket %s for terraform remote state\n", bucketName)
			//bucket var

			err := simulator.Create(logger, tfDir, bucketName, attackTag, tfVarsDir)
			if err != nil {
				logger.Errorw("Error creating infrastructure", zap.Error(err))
			}

			simulator := simulator.NewSimulator(
				simulator.WithLogger(logger),
				simulator.WithTfDir(tfDir),
				simulator.WithScenariosDir(scenariosDir),
				simulator.WithAttackTag(attackTag),
				simulator.WithBucketName(bucketName),
				simulator.WithTfVarsDir(tfVarsDir))

			cfg, err := simulator.SSHConfig()
			if err != nil {
				return errors.Wrap(err, "Error getting SSH config")
			}

			err = ssh.EnsureSSHConfig(*cfg)
			if err != nil {
				return errors.Wrapf(err, "Error writing SSH config")
			}

			return err
		},
	}

	return cmd
}

func newStatusCommand(logger *zap.SugaredLogger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   `status`,
		Short: "Gets the status of the infrastructure",
		RunE: func(cmd *cobra.Command, args []string) error {
			bucket := viper.GetString("state-bucket")
			if bucket == "" {
				logger.Warnf("Simulator has not been initialised with an S3 bucket")
				logger.Warn("Please run simulator init")
				return nil
			}
			tfDir := viper.GetString("tf-dir")
			tfVarsDir := viper.GetString("tf-vars-dir")
			attackTag := viper.GetString("attack-container-tag")
			tfo, err := simulator.Status(logger, tfDir, bucket, attackTag, tfVarsDir)
			if err != nil {
				logger.Errorw("Error getting status of infrastructure", zap.Error(err))
				return err
			}

			if tfo.BastionPublicIP.Value == "" {
				logger.Error("No Infrastructure found")
			} else {
				logger.Infof("Bastion IP: %s", tfo.BastionPublicIP.Value)
				logger.Infof("Master IPs: %v", tfo.MasterNodesPrivateIP.Value)
				logger.Infof("Cluster IPs: %v", tfo.ClusterNodesPrivateIP.Value)
			}

			return err
		},
	}

	return cmd
}

func newDestroyCommand(logger *zap.SugaredLogger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   `destroy`,
		Short: "Tears down the infrastructure created for scenarios",
		RunE: func(cmd *cobra.Command, args []string) error {
			bucket := viper.GetString("state-bucket")
			if bucket == "" {
				logger.Warnf("Simulator has not been initialised with an S3 bucket")
				logger.Warn("Please run simulator init")
				return nil
			}
			tfDir := viper.GetString("tf-dir")
			tfVarsDir := viper.GetString("tf-vars-dir")

			attackTag := viper.GetString("attack-container-tag")
			err := simulator.Destroy(logger, tfDir, bucket, attackTag, tfVarsDir)
			if err != nil {
				logger.Errorw("Error destroying infrastructure", zap.Error(err))
			}

			return err
		},
	}

	return cmd
}

func newInfraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           `infra <subcommand>`,
		Short:         "Interact with AWS to create, query and destroy the required infrastructure for scenarios",
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	logger, err := newLogger(viper.GetString("loglevel"), "console")
	if err != nil {
		logger.Fatalf("Can't re-initialize zap logger: %v", err)
	}

	cmd.AddCommand(newCreateCommand(logger))
	cmd.AddCommand(newStatusCommand(logger))
	cmd.AddCommand(newDestroyCommand(logger))

	return cmd
}
