package main

import (
	"math"
)

// Sum, or reduction, which computes the sum of the points in each cluster
// Divide each cluster sum by the number of points in that cluster
// Reassign, or map, the points to the cluster to the closest centroid

const RGBI = 3

func KmeansPartition(imgPix []uint8) [][3]float64 {
	var numClusters = 16
	threshold := 0.0001 //% objects change membership

	numObjs := len(imgPix)
	numPix := numObjs / 4

	// numClusters = numPix / 2

	// if numPix < numClusters {
	// 	numClusters = numPix - 1
	// }

	clusterSizes := make([]int, numClusters)
	membership := make([]int, numPix)
	clusters := make([][3]float64, numClusters)

	//random values
	// for i := 0; i < numClusters; i++ {
	// 	var startingPoint [3]uint8
	// 	for i := 0; i < RGBI; i++ {
	// 		startingPoint[i] = uint8(rand.Int())
	// 	}
	// 	clusters[i] = startingPoint
	// }

	//first n values
	for i := 0; i < numClusters; i++ {
		fixedIndex := i * 4
		p := imgPix[fixedIndex : fixedIndex+RGBI : fixedIndex+RGBI]
		point := [3]float64{float64(p[0]), float64(p[1]), float64(p[2])}
		clusters[i] = point
	}

	var (
		errorc        float64
		previousError float64

		// mu sync.Mutex // guards balance

	)
	for ok := true; ok; ok = math.Abs(errorc-previousError) >= threshold {
		errorc = 0.0
		newClusters := make([][3]float64, numClusters)
		newClusterSizes := make([]int, numClusters)

		for i := 0; i < numObjs; i += 4 {

			// go func(i int) {
			p := imgPix[i : i+RGBI : i+RGBI]
			point := [3]float64{float64(p[0]), float64(p[1]), float64(p[2])}

			index := nearestCluster(numClusters, point, clusters)
			fixedIndex := i / 4
			if membership[fixedIndex] != index {
				errorc += 1
			}

			membership[fixedIndex] = index

			//update new cluster center
			newClusterSizes[index]++
			for j := 0; j < 3; j++ {
				newClusters[index][j] += point[j]

			}
			// }(i)
		}

		//MAIN
		//average the sum and replace old cluster with new ones
		for i := 0; i < numClusters; i++ {
			size := newClusterSizes[i]
			if size > 0 {
				for j := 0; j < 3; j++ {
					clusters[i][j] = newClusters[i][j] / float64(size)
				}
			}
		}
		clusterSizes = newClusterSizes
		previousError = errorc
	}
	clustersInfo := TwoSlices{clusters, clusterSizes}
	// fmt.Println("", clusters)
	// fmt.Println("_____", clusterSizes)
	// sort.Sort(TwoSlices(clustersInfo))
	// fmt.Println("Sorted : ", clustersInfo.clusters, clustersInfo.clusterSizes)

	// w.Flush()
	// file.Close()

	// return clusters
	return clustersInfo.clusters
}

func nearestCluster(numClusters int, object [3]float64, clusters [][3]float64) int {
	index := 0
	min_dist := distance(object, clusters[0])
	for i := 0; i < numClusters; i++ {
		dist := distance(object, clusters[i])
		if dist < min_dist {
			min_dist = dist
			index = i
		}
	}
	return index
}
func distance(coord1 [3]float64, coord2 [3]float64) float64 {
	ans := 0.0
	for i := 0; i < 3; i++ {
		ans += float64(coord1[i]-coord2[i]) * float64(coord1[i]-coord2[i])
	}
	return ans
}

type TwoSlices struct {
	clusters     [][3]float64
	clusterSizes []int
}

func (sbo TwoSlices) Len() int {
	return len(sbo.clusters)
}
func (sbo TwoSlices) Swap(i, j int) {
	sbo.clusters[i], sbo.clusters[j] = sbo.clusters[j], sbo.clusters[i]
	sbo.clusterSizes[i], sbo.clusterSizes[j] = sbo.clusterSizes[j], sbo.clusterSizes[i]
}
func (clustersInfo TwoSlices) Less(i, j int) bool {
	return clustersInfo.clusterSizes[i] > clustersInfo.clusterSizes[j]
}
