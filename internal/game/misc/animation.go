// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package misc

import (
	"fmt"
	"time"

	"github.com/divVerent/aaaaxy/internal/animation"
	"github.com/divVerent/aaaaxy/internal/engine"
	"github.com/divVerent/aaaaxy/internal/level"
)

// Animation is a simple entity type that renders a static sprite. It can be optionally solid and/or opaque.
type Animation struct {
	SpriteBase
	Entity *engine.Entity
	Anim   animation.State
}

func (a *Animation) Spawn(w *engine.World, sp *level.SpawnableProps, e *engine.Entity) error {
	a.Entity = e
	prefix := sp.Properties["animation"]
	groupName := sp.Properties["animation_group"]
	group := &animation.Group{
		NextAnim: groupName,
	}
	framesString := sp.Properties["animation_frames"]
	if _, err := fmt.Sscanf(framesString, "%d", &group.Frames); err != nil {
		return fmt.Errorf("could not decode animation_frames %q: %v", framesString, err)
	}
	symmetricString := sp.Properties["animation_symmetric"]
	if symmetricString != "" {
		if _, err := fmt.Sscanf(symmetricString, "%t", &group.Symmetric); err != nil {
			return fmt.Errorf("could not decode animation_symmetric %q: %v", symmetricString, err)
		}
	}
	frameIntervalString := sp.Properties["animation_frame_interval"]
	if _, err := fmt.Sscanf(frameIntervalString, "%d", &group.FrameInterval); err != nil {
		return fmt.Errorf("could not decode animation_frame_interval %q: %v", frameIntervalString, err)
	}
	repeatIntervalString := sp.Properties["animation_repeat_interval"]
	if _, err := fmt.Sscanf(repeatIntervalString, "%d", &group.NextInterval); err != nil {
		return fmt.Errorf("could not decode animation_repeat_interval %q: %v", repeatIntervalString, err)
	}
	syncToMusicOffsetString := sp.Properties["animation_sync_to_music_offset"]
	if syncToMusicOffsetString != "" {
		var err error
		if group.SyncToMusicOffset, err = time.ParseDuration(syncToMusicOffsetString); err != nil {
			return fmt.Errorf("could not decode animation_sync_to_music_offset %q: %v", syncToMusicOffsetString, err)
		}
	}
	offsetString := sp.Properties["render_offset"]
	if offsetString == "" {
		e.ResizeImage = true
	} else {
		if _, err := fmt.Sscanf(offsetString, "%d %d", &e.RenderOffset.DX, &e.RenderOffset.DY); err != nil {
			return fmt.Errorf("could not decode render offset %q: %v", offsetString, err)
		}
	}
	if s := sp.Properties["border_pixels"]; s != "" {
		if _, err := fmt.Sscanf(s, "%d", &e.BorderPixels); err != nil {
			return fmt.Errorf("failed to decode borde pixels %q: %v", s, err)
		}
	}
	err := a.Anim.Init(prefix, map[string]*animation.Group{groupName: group}, groupName)
	if err != nil {
		return fmt.Errorf("could not initialize animation %v: %v", prefix, err)
	}
	return a.SpriteBase.Spawn(w, sp, e)
}

func (a *Animation) Update() {
	a.Anim.Update(a.Entity)
}

func init() {
	engine.RegisterEntityType(&Animation{})
}