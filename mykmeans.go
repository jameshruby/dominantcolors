package main

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

const RGBI = 3

//TODO replace 3 with shared const
type Cluster struct {
	Center [RGBI]byte
	Points [][RGBI]byte //Observations
}

func ResetClusters(clusters []Cluster) {
	for i := 0; i < len(clusters); i++ {
		clusters[i].Points = nil
	}
}

func RecenterClusters(clusters []Cluster) {
	for i := 0; i < len(clusters); i++ {
		clusters[i].Recenter()
	}
}
func (cluster *Cluster) Append(point [RGBI]byte) {
	cluster.Points = append(cluster.Points, point)
}
func (cluster *Cluster) Recenter() {
	var length = len(cluster.Points)
	var centerCoordinates [RGBI]byte
	for _, point := range cluster.Points {
		for j, v := range point {
			centerCoordinates[j] += v
		}
	}
	var mean [RGBI]byte
	for i, v := range centerCoordinates {
		mean[i] = byte(v / byte(length))
	}

	cluster.Center = mean
}

func NewCluster(clusterCount int) []Cluster {
	//TODO FIX err check
	var clusters []Cluster
	//TODO check size

	//We are looking at our points,
	//and pick N random starting points for our clusters
	// the only thing...they are COMPLETETELY random, which means possibly out of range...
	//TODO OPTIMIZE

	//generateStartingPoints
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < clusterCount; i++ {
		var startingPoint [RGBI]byte
		for i := 0; i < RGBI; i++ {
			startingPoint[i] = byte(rand.Int())
		}
		clusters = append(clusters, Cluster{Center: startingPoint})
	}

	return clusters
}
func NearestClusterIndex(clusters []Cluster, point [RGBI]byte) int {
	nearestDist := -1.0
	var nearestClusterIndex int
	for i, cluster := range clusters {
		currentDist := DistancePoints(cluster.Center, point)
		if nearestDist < 0 || currentDist < nearestDist {
			nearestDist = currentDist
			nearestClusterIndex = i
		}
	}
	return nearestClusterIndex
}

func DistancePoints(p1 [RGBI]byte, p2 [RGBI]byte) float64 {
	var r float64
	for i, v := range p1 {
		va := float64(v) - float64((p2[i]))
		r += math.Pow(float64(va), 2)
	}
	return r
}

//this version could iterate directly over imgPix
func KmeansPartition(imgPix []byte, clustersCount int) []Cluster {
	const iterationTreshold = 0
	const deltaThreshold = 0.01

	//TODO make clusters to accept [][RGBLEN] slice
	//instead of point structure
	clusters := NewCluster(clustersCount)
	changes := 1
	pixelCount := len(imgPix) / 4
	pointCenters := make([]int, pixelCount)

	for i := 0; changes > 0; i++ {
		changes = 0
		ResetClusters(clusters)

		var point [RGBI]byte
		z := 0
		for i := 0; i < len(imgPix); i += 4 {
			copy(point[:], imgPix[i:i+RGBI:i+RGBI])
			nearestClusterIndex := NearestClusterIndex(clusters, point)
			//add point to its nearest cluster
			clusters[nearestClusterIndex].Append(point)
			// check if the cluster for given point changed, or
			//whether the given cluster is staying the same
			if pointCenters[z] != nearestClusterIndex {
				pointCenters[z] = nearestClusterIndex
				changes++
			}
			z++
		}

		for clusterIndex := 0; clusterIndex < len(clusters); clusterIndex++ {
			if len(clusters[clusterIndex].Points) == 0 {
				// During the iterations, if any of the cluster centers has no
				// data points associated with it, assign a random data point
				// to it.
				// Also see: http://user.ceng.metu.edu.tr/~tcan/ceng465_f1314/Schedule/KMeansEmpty.html
				var randomIndex int
				for {
					// find a cluster with at least two data points, otherwise
					// we're just emptying one cluster to fill another
					randomIndex = rand.Intn(len(imgPix) - 4)
					if len(clusters[pointCenters[(randomIndex/4)]].Points) > 1 {
						break
					}
				}
				var point [RGBI]byte
				copy(point[:], imgPix[randomIndex:randomIndex+RGBI:randomIndex+RGBI])
				clusters[clusterIndex].Append(point)
				pointCenters[randomIndex/4] = clusterIndex

				// Ensure that we always see at least one more iteration after.
				// randomly assigning a data point to a cluster
				changes = pixelCount
			}
		}

		if changes > 0 {
			RecenterClusters(clusters)
		}
		if i == iterationTreshold ||
			changes < int(float64(pixelCount)*deltaThreshold) {
			// fmt.Println("Aborting:", changes, int(float64(len(dataset))*m.TerminationThreshold))
			break
		}
	}

	return clusters
}

func KmeansPartition2(imgPix []uint8) [][3]float64 {
	var numClusters = 16
	threshold := 0.0001 /* % objects change membership */

	numObjs := len(imgPix)
	numPix := numObjs / 4

	if numPix < numClusters {
		numClusters = numPix - 1
	}

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

	var errorc float64
	var previousError float64

	// file, err := os.Create("coords.txt")
	// if err != nil {
	// 	return nil
	// }
	// w := bufio.NewWriter(file)
	for ok := true; ok; ok = math.Abs(errorc-previousError) >= threshold {
		errorc = 0.0
		newClusters := make([][3]float64, numClusters)
		newClusterSizes := make([]int, numClusters)

		for i := 0; i < numObjs; i += 4 {
			p := imgPix[i : i+RGBI : i+RGBI]
			point := [3]float64{float64(p[0]), float64(p[1]), float64(p[2])}

			index := find_nearest_cluster2(numClusters, point, clusters)
			fixedIndex := i / 4
			if membership[fixedIndex] != index {
				errorc += 1
			}

			// space := "  "
			// if fixedIndex >= 10 {
			// 	space = " "
			// }
			// if fixedIndex >= 100 {
			// 	space = ""
			// }
			// fmt.Fprintln(w, fmt.Sprintf("%s%d %d %d %d", space, fixedIndex, point[0], point[1], point[2]))
			// _, err = w.WriteString("ttt") //fmt.Sprintf("%d %v \n", fixedIndex, point)
			// if err != nil {
			// 	return nil
			// }

			membership[fixedIndex] = index
			//update new cluster center
			newClusterSizes[index]++
			for j := 0; j < 3; j++ {
				newClusters[index][j] += point[j]
			}
		}

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
	sort.Sort(TwoSlices(clustersInfo))
	// fmt.Println("Sorted : ", clustersInfo.clusters, clustersInfo.clusterSizes)

	// w.Flush()
	// file.Close()

	// return clusters
	return clustersInfo.clusters
}
func euclid_dist_22(coord1 [3]float64, coord2 [3]float64) float64 {
	ans := 0.0
	for i := 0; i < 3; i++ {
		ans += float64(coord1[i]-coord2[i]) * float64(coord1[i]-coord2[i])
	}
	return ans
}
func find_nearest_cluster2(numClusters int, object [3]float64, clusters [][3]float64) int {
	index := 0
	min_dist := euclid_dist_22(object, clusters[0])
	for i := 0; i < numClusters; i++ {
		dist := euclid_dist_22(object, clusters[i])
		if dist < min_dist {
			min_dist = dist
			index = i
		}
	}
	return index
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
