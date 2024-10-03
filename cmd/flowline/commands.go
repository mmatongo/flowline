package main

import (
	"fmt"

	"github.com/mmatongo/flowline/outline"
	"github.com/mmatongo/flowline/pkg/logger"
	"github.com/spf13/cobra"
)

var log = logger.NewLogger()

var rootCmd = &cobra.Command{
	Use:   "flowline",
	Short: "Flowline is a util for migrating your confluence knowledge base...",
	Long:  "A tiny little tool built in a few hours out of frustration to migrate a confluence knowledge base",
}

var outlineCmd = &cobra.Command{
	Use:   "outline",
	Short: "Process a confluence HTML export and import it into Outline",
	Long:  "Process and convert individual pages to markdown, while aiming to preserve document structure",
	Run: func(cmd *cobra.Command, args []string) {
		inputDir, _ := cmd.Flags().GetString("input")
		outputDir, _ := cmd.Flags().GetString("output")
		collectionId, _ := cmd.Flags().GetString("collection")
		verify, _ := cmd.Flags().GetBool("verify")
		getCollections, _ := cmd.Flags().GetBool("get-collections")

		if getCollections {
			if inputDir == "" && outputDir == "" && collectionId == "" {
				res, err := outline.GetCollections(log)
				if err != nil {
					log.Logger.Error("Failed to retrieve collections: ", err)
					return
				}
				fmt.Println(res)
				return
			}
		}

		if inputDir == "" || outputDir == "" && collectionId == "" {
			err := cmd.Help()
			if err != nil {
				log.Logger.Error("failed to print help: ", err)
			}

			return
		}

		if inputDir != "" && outputDir != "" && collectionId != "" {
			if err := outline.PrepareAndProcess(inputDir, outputDir, collectionId, verify, log); err != nil {
				log.Logger.Error("failed to process confluence export ", err)
				return
			}
			log.Logger.Print("processing completed successfully")
		}
	},
}

func init() {
	outlineCmd.Flags().StringP("input", "i", "", "path to the confluence HTML export")
	outlineCmd.Flags().StringP("output", "o", "", "desired output path for the processed documents")
	outlineCmd.Flags().StringP("collection", "c", "", "collection id to be populated")
	outlineCmd.Flags().BoolP("get-collections", "G", false, "retrieve a list of all the collections")
	outlineCmd.Flags().BoolP("verify", "r", false, "verify the contents of each page before upload")

	rootCmd.AddCommand(outlineCmd)
}
