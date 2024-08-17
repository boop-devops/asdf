// Package cmd contains the asdf CLI command code
package cmd

import (
	"errors"
	"log"
	"os"

	"asdf/config"
	"asdf/internal/info"
	"asdf/plugins"

	"github.com/urfave/cli/v2"
)

const usageText = `The Multiple Runtime Version Manager.

Manage all your runtime versions with one tool!

Complete documentation is available at https://asdf-vm.com/`

// Execute defines the full CLI API and then runs it
func Execute(version string) {
	logger := log.New(os.Stderr, "", 0)
	log.SetFlags(0)

	app := &cli.App{
		Name:    "asdf",
		Version: "0.1.0",
		// Not really sure what I should put here, but all the new Golang code will
		// likely be written by me.
		Copyright: "(c) 2024 Trevor Brown",
		Authors: []*cli.Author{
			{
				Name: "Trevor Brown",
			},
		},
		Usage:     "The multiple runtime version manager",
		UsageText: usageText,
		Commands: []*cli.Command{
			{
				Name: "info",
				Action: func(_ *cli.Context) error {
					conf, err := config.LoadConfig()
					if err != nil {
						logger.Printf("error loading config: %s", err)
						return err
					}

					return infoCommand(conf, version)
				},
			},
			{
				Name: "plugin",
				Action: func(_ *cli.Context) error {
					logger.Println("Unknown command: `asdf plugin`")
					os.Exit(1)
					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name: "add",
						Action: func(cCtx *cli.Context) error {
							args := cCtx.Args()
							conf, err := config.LoadConfig()
							if err != nil {
								logger.Printf("error loading config: %s", err)
								return err
							}

							return pluginAddCommand(cCtx, conf, logger, args.Get(0), args.Get(1))
						},
					},
					{
						Name: "list",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "urls",
								Usage: "Show URLs",
							},
							&cli.BoolFlag{
								Name:  "refs",
								Usage: "Show Refs",
							},
						},
						Action: func(cCtx *cli.Context) error {
							return pluginListCommand(cCtx, logger)
						},
					},
					{
						Name: "remove",
						Action: func(cCtx *cli.Context) error {
							args := cCtx.Args()
							return pluginRemoveCommand(cCtx, logger, args.Get(0))
						},
					},
					{
						Name: "update",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "all",
								Usage: "Update all installed plugins",
							},
						},
						Action: func(cCtx *cli.Context) error {
							args := cCtx.Args()
							return pluginUpdateCommand(cCtx, logger, args.Get(0), args.Get(1))
						},
					},
				},
			},
		},
		Action: func(_ *cli.Context) error {
			// TODO: flesh this out
			log.Print("Late but latest -- Rajinikanth")
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}

func pluginAddCommand(_ *cli.Context, conf config.Config, logger *log.Logger, pluginName, pluginRepo string) error {
	if pluginName == "" {
		// Invalid arguments
		// Maybe one day switch this to show the generated help
		// cli.ShowSubcommandHelp(cCtx)
		return cli.Exit("usage: asdf plugin add <name> [<git-url>]", 1)
	}

	err := plugins.Add(conf, pluginName, pluginRepo)
	if err != nil {
		logger.Printf("%s", err)

		var existsErr plugins.PluginAlreadyExists
		if errors.As(err, &existsErr) {
			os.Exit(0)
			return nil
		}

		os.Exit(1)
		return nil
	}

	os.Exit(0)
	return nil
}

func pluginRemoveCommand(_ *cli.Context, logger *log.Logger, pluginName string) error {
	conf, err := config.LoadConfig()
	if err != nil {
		logger.Printf("error loading config: %s", err)
		return err
	}

	err = plugins.Remove(conf, pluginName)
	if err != nil {
		logger.Printf("error removing plugin: %s", err)
	}
	return err
}

func pluginListCommand(cCtx *cli.Context, logger *log.Logger) error {
	urls := cCtx.Bool("urls")
	refs := cCtx.Bool("refs")

	conf, err := config.LoadConfig()
	if err != nil {
		logger.Printf("error loading config: %s", err)
		return err
	}

	plugins, err := plugins.List(conf, urls, refs)
	if err != nil {
		logger.Printf("error loading plugin list: %s", err)
		return err
	}

	// TODO: Add some sort of presenter logic in another file so we
	// don't clutter up this cmd code with conditional presentation
	// logic
	for _, plugin := range plugins {
		if urls && refs {
			logger.Printf("%s\t\t%s\t%s\n", plugin.Name, plugin.URL, plugin.Ref)
		} else if refs {
			logger.Printf("%s\t\t%s\n", plugin.Name, plugin.Ref)
		} else if urls {
			logger.Printf("%s\t\t%s\n", plugin.Name, plugin.URL)
		} else {
			logger.Printf("%s\n", plugin.Name)
		}
	}

	return nil
}

func infoCommand(conf config.Config, version string) error {
	return info.Print(conf, version)
}

func pluginUpdateCommand(cCtx *cli.Context, logger *log.Logger, pluginName, ref string) error {
	updateAll := cCtx.Bool("all")
	if !updateAll && pluginName == "" {
		return cli.Exit("usage: asdf plugin-update {<name> [git-ref] | --all}", 1)
	}

	conf, err := config.LoadConfig()
	if err != nil {
		logger.Printf("error loading config: %s", err)
		return err
	}

	if updateAll {
		installedPlugins, err := plugins.List(conf, false, false)
		if err != nil {
			logger.Printf("failed to get plugin list: %s", err)
			return err
		}

		for _, plugin := range installedPlugins {
			updatedToRef, err := plugins.Update(conf, plugin.Name, "")
			formatUpdateResult(logger, plugin.Name, updatedToRef, err)
		}

		return nil
	}

	updatedToRef, err := plugins.Update(conf, pluginName, ref)
	formatUpdateResult(logger, pluginName, updatedToRef, err)
	return err
}

func formatUpdateResult(logger *log.Logger, pluginName, updatedToRef string, err error) {
	if err != nil {
		logger.Printf("failed to update %s due to error: %s\n", pluginName, err)

		return
	}

	logger.Printf("updated %s to ref %s\n", pluginName, updatedToRef)
}