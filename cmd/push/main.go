// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/slides/v1"
)

var (
	presentationID = flag.String("presentation", "", "ID of the presentation to push to")
)

func main() {
	flag.Parse()

	for i := range flag.NArg() {
		if err := createImageSlide(*presentationID, flag.Arg(i)); err != nil {
			log.Fatal(err)
		}
	}
}

func createImageSlide(presentationID, imageURL string) error {
	slideID, err := createSlide(presentationID)
	if err != nil {
		return err
	}
	return addImageToSlide(presentationID, slideID, imageURL)
}

func addImageToSlide(presentationID, slideID, imageURL string) error {
	slidesService := slidesClient()

	width := slides.Dimension{Magnitude: 10 * 914400 /*10 inches*/, Unit: "EMU"}
	height := slides.Dimension{Magnitude: 5.63 * 914400 /*5.63 inches*/, Unit: "EMU"}
	requests := []*slides.Request{{
		CreateImage: &slides.CreateImageRequest{
			Url: imageURL,
			ElementProperties: &slides.PageElementProperties{
				PageObjectId: slideID,
				Size: &slides.Size{
					Width:  &width,
					Height: &height,
				},
				Transform: &slides.AffineTransform{
					ScaleX:     1.0,
					ScaleY:     1.0,
					TranslateX: 0.0,
					TranslateY: 0.0,
					Unit:       "EMU",
				},
			},
		},
	}}

	// Execute the request.
	body := &slides.BatchUpdatePresentationRequest{
		Requests: requests,
	}
	_, err := slidesService.Presentations.BatchUpdate(presentationID, body).Do()
	if err != nil {
		return fmt.Errorf("failed to create image object: %v", err)
	}
	return nil
}

func createSlide(presentationID string) (string, error) {
	slidesService := slidesClient()

	// Add a slide at the end with BLANK layout.
	requests := []*slides.Request{{
		CreateSlide: &slides.CreateSlideRequest{
			SlideLayoutReference: &slides.LayoutReference{
				PredefinedLayout: "BLANK",
			},
		},
	}}

	// Execute the request.
	body := &slides.BatchUpdatePresentationRequest{
		Requests: requests,
	}
	response, err := slidesService.Presentations.BatchUpdate(presentationID, body).Do()
	if err != nil {
		return "", fmt.Errorf("failed to create slide: %v", err)
	}
	return response.Replies[0].CreateSlide.ObjectId, nil
}

func slidesClient() *slides.Service {
	ctx := context.Background()
	slidesService, err := slides.NewService(ctx, option.WithScopes(drive.DriveScope, slides.PresentationsScope))
	if err != nil {
		log.Fatalf("error creating Slides client: %v", err)
	}
	return slidesService
}
