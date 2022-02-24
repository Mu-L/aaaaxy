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

package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/divVerent/aaaaxy/internal/flag"
	"github.com/divVerent/aaaaxy/internal/log"
	"github.com/divVerent/aaaaxy/internal/shader"
)

var (
	drawBlurs = flag.Bool("draw_blurs", true, "perform blur effects; requires draw_visibility_mask")
)

func blurImageFixedFunction(img, tmp, out *ebiten.Image, size int, scale, darken float64) {
	opts := ebiten.DrawImageOptions{
		CompositeMode: ebiten.CompositeModeLighter,
		Filter:        ebiten.FilterNearest,
	}
	size++
	// Only power-of-two blurs look good with this approach, so let's scale down the blur as much as needed.
	for size&(size-1) != 0 {
		size--
	}
	src := img
	opts.ColorM.Scale(1, 1, 1, 0.5)
	for size > 1 {
		size /= 2
		tmp.Fill(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
		opts.CompositeMode = ebiten.CompositeModeCopy
		opts.GeoM.Reset()
		opts.GeoM.Translate(-float64(size), 0)
		tmp.DrawImage(src, &opts)
		opts.CompositeMode = ebiten.CompositeModeLighter
		opts.GeoM.Reset()
		opts.GeoM.Translate(float64(size), 0)
		tmp.DrawImage(src, &opts)
		src = out
		if size <= 1 {
			opts.ColorM.Scale(1, 1, 1, scale)
			opts.ColorM.Translate(-darken, -darken, -darken, 0)
		}
		out.Fill(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
		opts.CompositeMode = ebiten.CompositeModeCopy
		opts.GeoM.Reset()
		opts.GeoM.Translate(0, -float64(size))
		out.DrawImage(tmp, &opts)
		opts.CompositeMode = ebiten.CompositeModeLighter
		opts.GeoM.Reset()
		opts.GeoM.Translate(0, float64(size))
		out.DrawImage(tmp, &opts)
	}
}

func BlurExpandImage(img, tmp, out *ebiten.Image, blurSize, expandSize int, scale, darken float64) {
	// Blurring and expanding can be done in a single step by doing a regular blur then scaling up at the last step.
	if !*drawBlurs {
		blurSize = 0
	}
	size := blurSize + expandSize
	scale *= (2*float64(size) + 1) / (2*float64(blurSize) + 1)
	BlurImage(img, tmp, out, size, scale, darken, 1.0)
}

var (
	blurBroken = false
)

func BlurImage(img, tmp, out *ebiten.Image, size int, scale, darken, blurFade float64) {
	scale *= scale * blurFade
	scale += 1 - blurFade
	darken *= blurFade
	if !*drawBlurs && scale <= 1 {
		// Blurs can be globally turned off.
		if img == out {
			if scale == 1.0 && darken == 0.0 {
				return
			}
			options := &ebiten.DrawImageOptions{
				CompositeMode: ebiten.CompositeModeCopy,
				Filter:        ebiten.FilterNearest,
			}
			tmp.DrawImage(img, options)
			options.ColorM.Scale(scale, scale, scale, 1.0)
			options.ColorM.Translate(-darken, -darken, -darken, 0.0)
			out.DrawImage(tmp, options)
		} else {
			options := &ebiten.DrawImageOptions{
				CompositeMode: ebiten.CompositeModeCopy,
				Filter:        ebiten.FilterNearest,
			}
			options.ColorM.Scale(scale, scale, scale, 1.0)
			options.ColorM.Translate(-darken, -darken, -darken, 0.0)
			out.DrawImage(img, options)
		}
		return
	}
	if blurBroken {
		blurImageFixedFunction(img, tmp, out, size, scale, darken)
		return
	}
	// Too bad we can't have integer uniforms, so we need to templatize this
	// shader instead. Should be faster than having conditionals inside the
	// shader code.
	blurShader, err := shader.Load("blur.kage", map[string]string{
		"Size": fmt.Sprint(size),
	})
	if err != nil {
		log.Errorf("BROKEN RENDERER, WILL FALLBACK: could not load blur shader: %v", err)
		blurBroken = true
		blurImageFixedFunction(img, tmp, out, size, scale, darken)
		return
	}
	w, h := img.Size()
	centerScale := 1.0 / (2*float64(size)*blurFade + 1)
	otherScale := blurFade * centerScale
	tmp.DrawRectShader(w, h, blurShader, &ebiten.DrawRectShaderOptions{
		CompositeMode: ebiten.CompositeModeCopy,
		Uniforms: map[string]interface{}{
			"Step":        []float32{1 / float32(w), 0},
			"CenterScale": float32(centerScale),
			"OtherScale":  float32(otherScale),
			"Add":         []float32{float32(-darken), float32(-darken), float32(-darken), 0.0},
		},
		Images: [4]*ebiten.Image{
			img,
			nil,
			nil,
			nil,
		},
	})
	out.DrawRectShader(w, h, blurShader, &ebiten.DrawRectShaderOptions{
		CompositeMode: ebiten.CompositeModeCopy,
		Uniforms: map[string]interface{}{
			"Step":        []float32{0, 1 / float32(h)},
			"CenterScale": float32(centerScale * scale),
			"OtherScale":  float32(otherScale * scale),
			"Add":         []float32{float32(-darken), float32(-darken), float32(-darken), 0.0},
		},
		Images: [4]*ebiten.Image{
			tmp,
			nil,
			nil,
			nil,
		},
	})
}