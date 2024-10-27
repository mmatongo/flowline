package main

import (
	"fmt"

	"github.com/mmatongo/flowline/internal/markdown"
	"github.com/mmatongo/flowline/internal/outline"
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
					log.Logger.Error("failed to retrieve collections: ", err)
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
			log.Print("processing completed successfully")
		}
	},
}

var markdownCmd = &cobra.Command{
	Use:   "markdown",
	Short: "Convert Confluence HTML export to markdown files",
	Long: `Convert a Confluence HTML export to markdown files while preserving the document hierarchy.
Each page will be converted to an index.md file in its own directory, maintaining the original structure
with the attachments in a separate directory within the page directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputDir, _ := cmd.Flags().GetString("input")
		outputDir, _ := cmd.Flags().GetString("output")
		verify, _ := cmd.Flags().GetBool("verify")

		if inputDir == "" || outputDir == "" {
			err := cmd.Help()
			if err != nil {
				log.Logger.Error("failed to print help: ", err)
			}
			return
		}

		if err := markdown.ExportToMarkdown(inputDir, outputDir, verify, log); err != nil {
			log.Logger.Errorf("failed to convert confluence export: %v", err)
			return
		}

		log.Print("conversion completed successfully!")
	},
}

func init() {
	outlineCmd.Flags().StringP("input", "i", "", "path to the confluence HTML export")
	outlineCmd.Flags().StringP("output", "o", "", "desired output path for the processed documents")
	outlineCmd.Flags().StringP("collection", "c", "", "collection id to be populated")
	outlineCmd.Flags().BoolP("get-collections", "G", false, "retrieve a list of all the collections")
	outlineCmd.Flags().BoolP("verify", "r", false, "verify the contents of each page before upload")

	outlineCmd.MarkFlagRequired("input")
	outlineCmd.MarkFlagRequired("output")
	outlineCmd.MarkFlagRequired("collection")

	markdownCmd.Flags().StringP("input", "i", "", "path to the confluence HTML export")
	markdownCmd.Flags().StringP("output", "o", "", "output path for the markdown files")
	markdownCmd.Flags().BoolP("verify", "r", false, "verify before proceeding with conversion")

	markdownCmd.MarkFlagRequired("input")
	markdownCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(outlineCmd)
	rootCmd.AddCommand(markdownCmd)
}
