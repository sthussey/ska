/*
Copyright 2025 - Scott Hussey, Jerrod Early
*/

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sthussey/ska/sink/console"
	"github.com/sthussey/ska/source/fs"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "ska",
		Usage: "A tool for scaffolding repository or directory structures",
		Commands: []*cli.Command{
			{
				Name:  "graph",
				Usage: "Operations on directory graphs",
				Commands: []*cli.Command{
					{
						Name:  "build",
						Usage: "Build a graph from a directory",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "path",
								Aliases:  []string{"p"},
								Usage:    "Path to the directory to build the graph from",
								Required: true,
							},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							rootPath := cmd.String("path")

							root, err := fs.BuildGraph(rootPath)
							if err != nil {
								return fmt.Errorf("failed to build graph: %w", err)
							}

							fmt.Printf("Successfully built graph from %s\n", rootPath)
							fmt.Printf("Root node: %s (%s)\n", root.Key(), root.Type())

							return nil
						},
					},
					{
						Name:  "print",
						Usage: "Print the graph structure of a directory",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "path",
								Aliases:  []string{"p"},
								Usage:    "Path to the directory to print the graph for",
								Required: true,
							},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							rootPath := cmd.String("path")

							root, err := fs.BuildGraph(rootPath)
							if err != nil {
								return fmt.Errorf("failed to build graph: %w", err)
							}

							console.PrintGraph(root, 0)
							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
