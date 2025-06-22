package main

import (
	"fmt"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 800
	screenHeight = 600
	boxSize      = 50
	trainHeadSize = 20
	cargoSize     = 20
	segmentGap    = 5
	trainSpeed    = 3.0
	railWidth     = 10
	refillTime    = 180 // Frames to wait before refilling a bottom box (3 seconds at 60 FPS)
	stationDelay  = 300 // Frames to wait at a station (5 seconds at 60 FPS)
)

type Box struct {
	rect        raylib.Rectangle
	color       raylib.Color
	isEmpty     bool
	refillTimer int // Timer for refilling empty bottom boxes
}

type Train struct {
	headPos            raylib.Vector2
	bodySegments       []raylib.Vector2
	cargo              [3]raylib.Color
	cargoCount         int
	path               []raylib.Vector2
	currentWaypointIdx int
	targetPos          raylib.Vector2
	isMoving           bool
	isFilling          bool
	isEmptying         bool
	stationTimer       int
}

var (
	topBoxes    [3]Box
	bottomBoxes [3]Box
	train       Train
	railPoints  []raylib.Vector2
)

func initGame() {
	// Initialize top boxes (empty)
	for i := 0; i < 3; i++ {
		topBoxes[i] = Box{
			rect:    raylib.NewRectangle(screenWidth/2-boxSize*1.5+float32(i)*boxSize, screenHeight/4-boxSize/2, boxSize, boxSize),
			color:   raylib.LightGray,
			isEmpty: true,
		}
	}

	// Initialize bottom boxes (black)
	for i := 0; i < 3; i++ {
		bottomBoxes[i] = Box{
			rect:        raylib.NewRectangle(screenWidth/2-boxSize*1.5+float32(i)*boxSize, screenHeight*3/4-boxSize/2, boxSize, boxSize),
			color:       raylib.Black,
			isEmpty:     false,
			refillTimer: 0, // No timer needed initially as they are full
		}
	}

	// Define the complex rail path
	railPoints = []raylib.Vector2{
		// Path 1: Middle bottom center to middle top center (via left)
		{X: screenWidth / 2, Y: screenHeight*3/4 + boxSize/2 + trainHeadSize}, // Start at bottom center
		{X: screenWidth / 2, Y: screenHeight*3/4 + boxSize/2 + trainHeadSize}, // Stay at bottom center for a moment
		{X: screenWidth/2 - 150, Y: screenHeight*3/4 + boxSize/2 + trainHeadSize}, // Move left
		{X: screenWidth/2 - 150, Y: screenHeight/4 - boxSize/2 - trainHeadSize}, // Move up
		{X: screenWidth / 2, Y: screenHeight/4 - boxSize/2 - trainHeadSize},   // Move to top center

		// Path 2: Middle top center to middle bottom center (via right)
		{X: screenWidth / 2, Y: screenHeight/4 - boxSize/2 - trainHeadSize},   // Stay at top center for a moment
		{X: screenWidth/2 + 150, Y: screenHeight/4 - boxSize/2 - trainHeadSize}, // Move right
		{X: screenWidth/2 + 150, Y: screenHeight*3/4 + boxSize/2 + trainHeadSize}, // Move down
		{X: screenWidth / 2, Y: screenHeight*3/4 + boxSize/2 + trainHeadSize}, // Move to bottom center
	}

	// Initialize train
	train = Train{
		headPos:            railPoints[0],
		bodySegments:       make([]raylib.Vector2, 3), // 3 cargo segments
		cargo:              [3]raylib.Color{raylib.Blank, raylib.Blank, raylib.Blank},
		cargoCount:         0,
		path:               railPoints,
		currentWaypointIdx: 0,
		targetPos:          railPoints[1], // Start by moving to the first actual movement point
		isMoving:           true,
		isFilling:          false,
		isEmptying:         false,
		stationTimer:       0,
	}

	// Initialize body segments behind the head
	for i := 0; i < 3; i++ {
		train.bodySegments[i] = raylib.NewVector2(train.headPos.X, train.headPos.Y+float32(i+1)*(cargoSize+segmentGap))
	}
}

func updateGame() {
	// Refill bottom boxes
	for i := 0; i < 3; i++ {
		if bottomBoxes[i].isEmpty {
			bottomBoxes[i].refillTimer--
			if bottomBoxes[i].refillTimer <= 0 {
				bottomBoxes[i].color = raylib.Black
				bottomBoxes[i].isEmpty = false
				bottomBoxes[i].refillTimer = 0 // Reset timer
			}
		}
	}

	if train.isMoving {
		// Move head towards target waypoint
		direction := raylib.Vector2Subtract(train.targetPos, train.headPos)
		distance := raylib.Vector2Length(direction)

		if distance > trainSpeed {
			direction = raylib.Vector2Scale(raylib.Vector2Normalize(direction), trainSpeed)
			train.headPos = raylib.Vector2Add(train.headPos, direction)
		} else {
			train.headPos = train.targetPos
			train.currentWaypointIdx = (train.currentWaypointIdx + 1) % len(train.path)
			train.targetPos = train.path[train.currentWaypointIdx]

			// Check if train is at a "station" (near top or bottom boxes)
			// Bottom station (filling)
			if train.currentWaypointIdx == 1 || train.currentWaypointIdx == 8 { // Waypoint 1 and 8 are the "stay" points at bottom center
				if train.cargoCount == 0 && !train.isFilling {
					train.isMoving = false
					train.isFilling = true
					train.stationTimer = stationDelay // Start station timer
					fmt.Println("Train arrived at bottom station, starting to fill cargo.")
				}
			}
			// Top station (emptying)
			if train.currentWaypointIdx == 4 || train.currentWaypointIdx == 5 { // Waypoint 4 and 5 are the "stay" points at top center
				if train.cargoCount > 0 && !train.isEmptying {
					train.isMoving = false
					train.isEmptying = true
					train.stationTimer = stationDelay // Start station timer
					fmt.Println("Train arrived at top station, starting to empty cargo.")
				}
			}
		}

		// Make body segments follow the head
		// This is a simplified snake-like movement. For true curve following,
		// you'd need to store a history of head positions.
		// For now, let's just make them follow the previous segment's position.
		if len(train.bodySegments) > 0 {
			train.bodySegments[0] = train.headPos
			for i := 1; i < len(train.bodySegments); i++ {
				// Simple follow: each segment moves towards the previous one
				prevSegmentPos := train.bodySegments[i-1]
				currentSegmentPos := train.bodySegments[i]

				segDirection := raylib.Vector2Subtract(prevSegmentPos, currentSegmentPos)
				segDistance := raylib.Vector2Length(segDirection)

				if segDistance > float32(cargoSize+segmentGap) {
					segDirection = raylib.Vector2Scale(raylib.Vector2Normalize(segDirection), float32(trainSpeed))
					train.bodySegments[i] = raylib.Vector2Add(currentSegmentPos, segDirection)
				} else {
					// If too close, just snap to a position behind the previous segment
					// This is a very basic "snake" movement.
					train.bodySegments[i] = raylib.Vector2Subtract(prevSegmentPos, raylib.Vector2Scale(raylib.Vector2Normalize(segDirection), float32(cargoSize+segmentGap)))
				}
			}
		}

	} else { // Train is stopped for filling/emptying
		train.stationTimer--
		if train.stationTimer <= 0 {
			if train.isFilling {
				// Fill cargo from bottom boxes
				filledOne := false
				for i := 0; i < 3; i++ {
					if !bottomBoxes[i].isEmpty && train.cargo[i] == raylib.Blank {
						train.cargo[i] = bottomBoxes[i].color
						bottomBoxes[i].color = raylib.LightGray // Make bottom box empty visually
						bottomBoxes[i].isEmpty = true
						bottomBoxes[i].refillTimer = refillTime // Start refill timer
						train.cargoCount++
						filledOne = true
						break // Fill one cargo at a time for visual effect
					}
				}
				if !filledOne || train.cargoCount == 3 { // All cargo filled or no more black boxes
					train.isFilling = false
					train.isMoving = true
					fmt.Println("Train finished filling cargo. Cargo count:", train.cargoCount)
				} else {
					train.stationTimer = stationDelay // Reset timer for next cargo
				}
			} else if train.isEmptying {
				// Empty cargo to top boxes
				emptiedOne := false
				for i := 0; i < 3; i++ {
					if topBoxes[i].isEmpty && train.cargo[i] != raylib.Blank {
						topBoxes[i].color = train.cargo[i]
						topBoxes[i].isEmpty = false
						train.cargo[i] = raylib.Blank // Empty train cargo slot
						train.cargoCount--
						emptiedOne = true
						break // Empty one cargo at a time for visual effect
					}
				}
				if !emptiedOne || train.cargoCount == 0 { // All cargo emptied or no more cargo
					train.isEmptying = false
					train.isMoving = true
					fmt.Println("Train finished emptying cargo. Cargo count:", train.cargoCount)
					// After emptying, make top boxes blank again for the next cycle
					for i := 0; i < 3; i++ {
						topBoxes[i].color = raylib.LightGray
						topBoxes[i].isEmpty = true
					}
				} else {
					train.stationTimer = stationDelay // Reset timer for next cargo
				}
			}
		}
	}
}

func drawGame() {
	raylib.BeginDrawing()
	raylib.ClearBackground(raylib.RayWhite)

	// Draw rail (draw lines between rail points)
	for i := 0; i < len(railPoints)-1; i++ {
		raylib.DrawLineEx(railPoints[i], railPoints[i+1], float32(railWidth), raylib.DarkGray)
	}

	// Draw top boxes
	for _, box := range topBoxes {
		raylib.DrawRectangleRec(box.rect, box.color)
		raylib.DrawRectangleLinesEx(box.rect, 2, raylib.Black)
	}

	// Draw bottom boxes
	for _, box := range bottomBoxes {
		raylib.DrawRectangleRec(box.rect, box.color)
		raylib.DrawRectangleLinesEx(box.rect, 2, raylib.Black)
	}

	// Draw train head
	raylib.DrawCircleV(train.headPos, float32(trainHeadSize)/2, raylib.Brown)
	raylib.DrawCircleLines(int32(train.headPos.X), int32(train.headPos.Y), float32(trainHeadSize)/2, raylib.DarkBrown)

	// Draw train cargo segments
	for i, segmentPos := range train.bodySegments {
		cargoRect := raylib.NewRectangle(segmentPos.X-float32(cargoSize)/2, segmentPos.Y-float32(cargoSize)/2, cargoSize, cargoSize)
		raylib.DrawRectangleRec(cargoRect, train.cargo[i])
		raylib.DrawRectangleLinesEx(cargoRect, 1, raylib.DarkGray)
	}

	raylib.EndDrawing()
}

func main() {
	raylib.InitWindow(screenWidth, screenHeight, "Train Automation Game")
	defer raylib.CloseWindow()

	raylib.SetTargetFPS(60)

	initGame()

	for !raylib.WindowShouldClose() {
		updateGame()
		drawGame()
	}
}
