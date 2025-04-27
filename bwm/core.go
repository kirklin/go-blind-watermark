// Copyright (c) 2025 Kirk Lin
// SPDX-License-Identifier: MIT

package bwm

import (
	"math"
	"sync"

	"gonum.org/v1/gonum/mat"
)

// processBlockEmbed 核心单块嵌入算法
// DCT -> SVD -> 修改奇异值 -> 逆SVD -> 逆DCT
func (b *BlindWatermark) processBlockEmbed(block *mat.Dense, bit bool) *mat.Dense {
	// 1. DCT
	dctBlock := PerformDCT(block)

	// 2. SVD
	var svd mat.SVD
	ok := svd.Factorize(dctBlock, mat.SVDThin)
	if !ok {
		return block // 如果 SVD 失败，返回原块（极其罕见）
	}

	// 获取 U, S, V
	var s []float64
	s = svd.Values(nil)
	var u, v mat.Dense
	svd.UTo(&u)
	svd.VTo(&v)

	// 3. 嵌入逻辑: 修改最大的奇异值 s[0]
	// 量化步长 d1
	val := s[0]
	bitVal := 0.0
	if bit {
		bitVal = 1.0
	}

	// 量化公式
	newVal := (math.Floor(val/b.D1) + 0.25 + 0.5*bitVal) * b.D1
	s[0] = newVal

	// 4. 重构 DCT 块 (U * Sigma * V^T)
	sigma := mat.NewDense(4, 4, nil)
	for k := 0; k < 4; k++ {
		if k < len(s) {
			sigma.Set(k, k, s[k])
		}
	}

	var temp, reconstructedDCT mat.Dense
	temp.Mul(&u, sigma)
	reconstructedDCT.Mul(&temp, v.T())

	// 5. 逆 DCT
	return PerformIDCT(&reconstructedDCT)
}

// processBlockExtract 核心单块提取算法
func (b *BlindWatermark) processBlockExtract(block *mat.Dense) bool {
	// 1. DCT
	dctBlock := PerformDCT(block)

	// 2. SVD
	var svd mat.SVD
	ok := svd.Factorize(dctBlock, mat.SVDThin)
	if !ok {
		return false
	}
	s := svd.Values(nil)

	// 3. 提取逻辑
	// 检查 s[0] 在量化区间的位置
	val := s[0]
	remainder := math.Mod(val, b.D1)
	return remainder > (b.D1 / 2.0)
}

// EmbedProcess 执行完整的嵌入流水线
func (b *BlindWatermark) EmbedProcess(imgData *YUVData, wmBits []bool) *YUVData {
	// 1. DWT 变换 Y 通道
	cA, cH, cV, cD := HaarDWT2D(imgData.Y)
	r, c := cA.Dims()

	// 2. 准备分块 (4x4)
	blockR, blockC := 4, 4
	rowsBlocks := r / blockR
	colsBlocks := c / blockC
	totalBlocks := rowsBlocks * colsBlocks

	// 3. 乱序索引
	indices := GenerateShuffledIndices(totalBlocks, b.SeedImg)
	// 打乱水印以便循环嵌入
	wmLen := len(wmBits)
	shuffledWmBits := make([]bool, wmLen)
	wmIndices := GenerateShuffledIndices(wmLen, b.SeedWm)
	for i, idx := range wmIndices {
		shuffledWmBits[i] = wmBits[idx]
	}

	// 4. 并发处理
	var wg sync.WaitGroup
	resultBlocks := make(map[int]*mat.Dense)
	var mu sync.Mutex

	// 为了减少锁竞争，我们可以分批或者直接用 slice，这里用 map 演示逻辑简单性
	// 实际高性能场景建议使用预分配的 slice

	// 这里将 cA 切分成小块的逻辑
	// 简化起见，我们直接遍历块索引
	for i := 0; i < totalBlocks; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// 计算真实的块坐标
			realPos := indices[idx]
			rowIdx := realPos / colsBlocks
			colIdx := realPos % colsBlocks

			// 提取 4x4 块
			block := cA.Slice(rowIdx*blockR, (rowIdx+1)*blockR, colIdx*blockC, (colIdx+1)*blockC).(*mat.Dense)

			// 获取对应的水印位 (循环使用)
			bit := shuffledWmBits[idx%wmLen]

			// 处理
			res := b.processBlockEmbed(block, bit)

			mu.Lock()
			resultBlocks[realPos] = res
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	// 5. 将处理完的块写回 cA
	for idx, block := range resultBlocks {
		rowIdx := idx / colsBlocks
		colIdx := idx % colsBlocks

		// Copy block back to cA
		// Gonum 没有直接的 SubMatrix Set，需要手动循环
		for i := 0; i < blockR; i++ {
			for j := 0; j < blockC; j++ {
				cA.Set(rowIdx*blockR+i, colIdx*blockC+j, block.At(i, j))
			}
		}
	}

	// 6. 逆 DWT
	newY := HaarIDWT2D(cA, cH, cV, cD)

	return &YUVData{Y: newY, U: imgData.U, V: imgData.V, W: imgData.W, H: imgData.H}
}

// ExtractProcess 执行完整的提取流水线
func (b *BlindWatermark) ExtractProcess(imgData *YUVData, wmLen int) []bool {
	cA, _, _, _ := HaarDWT2D(imgData.Y)
	r, c := cA.Dims()

	blockR, blockC := 4, 4
	rowsBlocks := r / blockR
	colsBlocks := c / blockC
	totalBlocks := rowsBlocks * colsBlocks

	indices := GenerateShuffledIndices(totalBlocks, b.SeedImg)

	extractedBitsRaw := make([]bool, totalBlocks)

	var wg sync.WaitGroup

	for i := 0; i < totalBlocks; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			realPos := indices[idx]
			rowIdx := realPos / colsBlocks
			colIdx := realPos % colsBlocks

			block := cA.Slice(rowIdx*blockR, (rowIdx+1)*blockR, colIdx*blockC, (colIdx+1)*blockC).(*mat.Dense)

			extractedBitsRaw[idx] = b.processBlockExtract(block)
		}(i)
	}
	wg.Wait()

	// 7. 聚合与反解密水印
	// 由于是循环嵌入，我们需要对提取出的结果进行平均或投票
	// 这里简化实现：假设水印长度已知，取余数对应的所有位置进行投票

	finalBits := make([]bool, wmLen)
	wmIndices := GenerateShuffledIndices(wmLen, b.SeedWm)

	// 投票桶：正票数, 总票数
	votes := make([]int, wmLen)
	counts := make([]int, wmLen)

	for i, bit := range extractedBitsRaw {
		wmPos := i % wmLen
		counts[wmPos]++
		if bit {
			votes[wmPos]++
		}
	}

	// 还原乱序并判定结果
	// 如果 True 的比例 > 0.5，则为 True
	shuffledResult := make([]bool, wmLen)
	for i := 0; i < wmLen; i++ {
		if float64(votes[i])/float64(counts[i]) > 0.5 {
			shuffledResult[i] = true
		} else {
			shuffledResult[i] = false
		}
	}

	// 反转水印的 Shuffle
	// wmIndices[original_pos] = shuffled_pos
	// 所以 finalBits[wmIndices[k]] = shuffledResult[k] ? No, Shuffle logic needs care.
	// Shuffle: arr[i] moves to indices[i] or source is indices[i]?
	// Utils shuffle: indices[0] is the index of the element that goes to position 0.
	// So: shuffled[i] comes from source[indices[i]]

	for i, sourceIdx := range wmIndices {
		finalBits[sourceIdx] = shuffledResult[i]
	}

	return finalBits
}
