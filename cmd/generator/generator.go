package generator

import (
	"fmt"
	"os"

	"github.com/lumina-tech/gooq/pkg/generator/templates"
	"github.com/spf13/viper"

	"github.com/jmoiron/sqlx"
	"github.com/lumina-tech/gooq/pkg/database"
	"github.com/lumina-tech/gooq/pkg/generator"
	"github.com/spf13/cobra"
)

var (
	generateDatabaseModelCommandUseDocker bool
	generateDatabaseModelConfigFilePath   string
)

var generateDatabaseModelCommand = &cobra.Command{
	Use:   "generate-database-model",
	Short: "generate Go models by introspecting the database",
	Run: func(cmd *cobra.Command, args []string) {
		err := initConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read configuration file:", err)
			os.Exit(1)
		}
		config := database.DatabaseConfig{
			Host:          viper.GetString("host"),
			Port:          viper.GetInt64("port"),
			Username:      viper.GetString("username"),
			Password:      viper.GetString("password"),
			DatabaseName:  viper.GetString("databaseName"),
			SSLMode:       viper.GetString("sslmode"),
			MigrationPath: viper.GetString("migrationPath"),
			ModelPath:     viper.GetString("modelPath"),
			TablePath:     viper.GetString("tablePath"),
		}
		if generateDatabaseModelCommandUseDocker {
			db := database.NewDockerizedDB(&config, viper.GetString("dockerTag"))
			defer db.Close()
			database.MigrateDatabase(db.DB.DB, config.MigrationPath)
			generateModelsForDB(db.DB, &config)
		} else {
			db := database.NewDatabase(&config)
			generateModelsForDB(db, &config)
		}
	},
}

func initConfig() error {
	viper.SetDefault("dockerTag", "11.4-alpine")

	if len(generateDatabaseModelConfigFilePath) != 0 {
		viper.SetConfigFile(generateDatabaseModelConfigFilePath)
		return viper.ReadInConfig()
	}

	viper.SetConfigName("gooq")

	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	viper.AddConfigPath(wd)

	return viper.ReadInConfig()
}

func generateModelsForDB(
	db *sqlx.DB, config *database.DatabaseConfig,
) {
	generator.NewEnumGenerator(db, templates.EnumTemplate, config.ModelPath, config.DatabaseName).Run()

	generatedModelFilename := fmt.Sprintf("%s/%s_model.generated.go", config.ModelPath, config.DatabaseName)
	generator.GenerateModel(db, templates.ModelTemplate, generatedModelFilename, config.DatabaseName, "model")

	generatedTableFilename := fmt.Sprintf("%s/%s_table.generated.go", config.TablePath, config.DatabaseName)
	generator.GenerateModel(db, templates.TableTemplate, generatedTableFilename, config.DatabaseName, "table")
}
