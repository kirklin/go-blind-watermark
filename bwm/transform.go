// Copyright (c) 2025 Kirk Lin
// SPDX-License-Identifier: MIT

package bwm

import (
	"gonum.org/v1/gonum/mat"
	"math"
)

// DCTMatrix4x4 预计算 4x4 DCT 变换矩阵
var DCTMatrix4x4 *mat.Dense
var IDCTMatrix4x4 *mat.Dense

func init() {
	// 初始化 DCT 矩阵
	t := make([]float64, 16)
	for i := 0; i < 4; i++ {
		alpha := 0.0
		if i == 0 {
			alpha = 1.0 / math.Sqrt(4.0)
		} else {
			alpha = math.Sqrt(2.0 / 4.0)
		}
		for j := 0; j < 4; j++ {
			t[i*4+j] = alpha * math.Cos((math.Pi*(2*float64(j)+1)*float64(i))/(2*4.0))
		}
	}
	DCTMatrix4x4 = mat.NewDense(4, 4, t)
	IDCTMatrix4x4 = mat.NewDense(4, 4, nil)
	IDCTMatrix4x4.CloneFrom(DCTMatrix4x4.T())
}

// HaarDWT2D 执行一级二维 Haar 小波变换
// 返回: cA (近似), cH (水平), cV (垂直), cD (对角)
func HaarDWT2D(data *mat.Dense) (*mat.Dense, *mat.Dense, *mat.Dense, *mat.Dense) {
	r, c := data.Dims()
	halfR, halfC := r/2, c/2

	cA := mat.NewDense(halfR, halfC, nil)
	cH := mat.NewDense(halfR, halfC, nil)
	cV := mat.NewDense(halfR, halfC, nil)
	cD := mat.NewDense(halfR, halfC, nil)

	// 简单的 Haar 实现：行变换 -> 列变换
	// 为了性能，这里合并步骤，直接计算
	for i := 0; i < halfR; i++ {
		for j := 0; j < halfC; j++ {
			// 获取 2x2 块
			// p1 p2
			// p3 p4
			r1, c1 := i*2, j*2
			p1 := data.At(r1, c1)
			p2 := data.At(r1, c1+1)
			p3 := data.At(r1+1, c1)
			p4 := data.At(r1+1, c1+1)

			// Haar logic
			cA.Set(i, j, (p1+p2+p3+p4)*0.5)
			cH.Set(i, j, (p1-p2+p3-p4)*0.5)
			cV.Set(i, j, (p1+p2-p3-p4)*0.5)
			cD.Set(i, j, (p1-p2-p3+p4)*0.5)
		}
	}
	return cA, cH, cV, cD
}

// HaarIDWT2D 执行一级二维 Haar 小波逆变换
func HaarIDWT2D(cA, cH, cV, cD *mat.Dense) *mat.Dense {
	r, c := cA.Dims()
	fullR, fullC := r*2, c*2
	out := mat.NewDense(fullR, fullC, nil)

	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			a := cA.At(i, j)
			h := cH.At(i, j)
			v := cV.At(i, j)
			d := cD.At(i, j)

			// Inverse logic
			out.Set(i*2, j*2, (a+h+v+d)*0.5)
			out.Set(i*2, j*2+1, (a-h+v-d)*0.5)
			out.Set(i*2+1, j*2, (a+h-v-d)*0.5)
			out.Set(i*2+1, j*2+1, (a-h-v+d)*0.5)
		}
	}
	return out
}

// PerformDCT 对 4x4 块进行 DCT
func PerformDCT(block *mat.Dense) *mat.Dense {
	// T * B * T'
	var temp mat.Dense
	temp.Mul(DCTMatrix4x4, block)
	var res mat.Dense
	res.Mul(&temp, IDCTMatrix4x4) // IDCTMatrix is T'
	return &res
}

// PerformIDCT 对 4x4 块进行逆 DCT
func PerformIDCT(block *mat.Dense) *mat.Dense {
	// T' * B * T
	var temp mat.Dense
	temp.Mul(IDCTMatrix4x4, block)
	var res mat.Dense
	res.Mul(&temp, DCTMatrix4x4)
	return &res
}
