# go-blind-watermark


> **A secure, robust, and pure Go blind watermarking library.**
>
> 图片盲水印：基于频域算法 (DWT + DCT + SVD)，抗压缩、抗裁剪，纯 Go 实现（无 CGO/OpenCV 依赖）。

## Introduction

**go-blind-watermark** (package `bwm`) is a high-performance library for embedding invisible watermarks into images. Unlike traditional visible watermarks, blind watermarks are embedded into the frequency domain of the image data. They remain invisible to the naked eye but can be algorithmically extracted even after the image has been compressed, cropped, or rotated.

### Key Features

* **Pure Go Implementation**: No dependencies on `OpenCV`, `numpy`, or `CGO`. Just pure Go and `Gonum`. Cross-compilation friendly.
* **High Performance**: Utilizes Go's concurrency model (Goroutines) to process image blocks in parallel.
* **Robustness**: Resistant to common attacks like JPEG compression, scaling, cropping, and noise addition.
* **Secure**: Watermarks are scrambled using a seed-based shuffling algorithm. Only the correct seed can extract the watermark.
* **Flexible**: Supports both **Text** and **Image (Logo)** watermarks.

## Mathematical Theory

This library implements a hybrid embedding scheme combining **DWT** (Discrete Wavelet Transform), **DCT** (Discrete Cosine Transform), and **SVD** (Singular Value Decomposition).

### 1\. The Pipeline

$$\text{Image} \xrightarrow{\text{YUV}} \text{Y-Channel} \xrightarrow{\text{DWT}} \text{LL Sub-band} \xrightarrow{\text{Block Split}} \text{4x4 Blocks}$$

Each 4x4 block then undergoes:
$$\text{Block} \xrightarrow{\text{DCT}} \text{Freq Coeffs} \xrightarrow{\text{SVD}} U \Sigma V^T$$

### 2\. Discrete Wavelet Transform (DWT)

We use the **Haar Wavelet** to decompose the image into four sub-bands: $LL$ (Approximation), $LH$ (Horizontal), $HL$ (Vertical), and $HH$ (Diagonal). The watermark is embedded in the $LL$ band to ensure stability against compression.

### 3\. Singular Value Decomposition (SVD)

The core embedding happens in the Singular Value matrix $\Sigma$. For a block $A$ in the DCT domain:

$$A = U \Sigma V^T$$

Where $\Sigma$ is a diagonal matrix containing singular values $\sigma_i$. The watermark bit $w \in \{0, 1\}$ is embedded by quantizing the largest singular value $\sigma_0$.

**Embedding Rule:**
$$\sigma'_0 = \left( \lfloor \frac{\sigma_0}{D} \rfloor + 0.25 + 0.5 \times w \right) \times D$$

Where $D$ is the quantization step (strength factor). A larger $D$ yields higher robustness but slightly lower image quality.

## 📦 Installation

```bash
go get github.com/kirklin/go-blind-watermark
```

## 🚀 Usage

### 1\. Embed Text Watermark

```go
package main

import (
	"image/png"
	"os"
	"github.com/kirklin/go-blind-watermark/bwm"
)

func main() {
	// 1. Load Source Image
	file, _ := os.Open("source.png")
	defer file.Close()
	srcImg, _, _ := image.Decode(file)

	// 2. Prepare Watermark
	text := "Copyright © Kirk Lin"
	wmBits := bwm.TextToBits(text)

	// 3. Initialize Engine (Seed1: Image Shuffle, Seed2: Watermark Encrypt)
	engine := bwm.New(2025, 8888)
	engine.D1 = 36.0 // Strength factor

	// 4. Embed
	outputImg, _ := engine.Embed(srcImg, wmBits)

	// 5. Save
	out, _ := os.Create("watermarked.png")
	defer out.Close()
	png.Encode(out, outputImg)
}
```

### 2\. Extract Watermark

```go
func main() {
    // ... Load watermarked image as `img` ...

    engine := bwm.New(2025, 8888)
    
    // You need to know the approximate length, or extract a fixed length
    extractLen := 128 
    
    bits, _ := engine.Extract(img, extractLen)
    
    text := bwm.BitsToText(bits)
    println("Extracted:", text)
}
```

### 3\. Embed Logo Image

For higher visual robustness, you can embed a binary logo instead of text.

```go
// Resize logo to fit capacity (e.g., 64x64)
logoImg := bwm.ResizeImage(originalLogo, 64, 64)

// Convert logo to bits (threshold 128)
wmBits, w, h := bwm.LogoToBits(logoImg, 128)

// Embed
engine.Embed(hostImg, wmBits)

// ...

// Extract and Restore
extractedBits, _ := engine.Extract(watermarkedImg, w*h)
recoveredLogo := bwm.BitsToLogo(extractedBits, w, h)
```

## License

Distributed under the MIT License. See `LICENSE` for more information.

-----

**Copyright © 2025 Kirk Lin**