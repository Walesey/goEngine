package editor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Invictus321/invictus321-countdown"
	"github.com/walesey/go-engine/assets"
	"github.com/walesey/go-engine/editor/models"
	"github.com/walesey/go-engine/renderer"
)

func (e *Editor) saveMap(filepath string) {
	data, err := json.Marshal(e.currentMap)
	if err != nil {
		log.Printf("Error Marshaling map file: %v\n", err)
		return
	}

	ioutil.WriteFile(filepath, data, os.ModePerm)
}

func (e *Editor) loadMap(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Error Reading map file: %v\n", err)
		return
	}

	var mapModel editorModels.MapModel
	err = json.Unmarshal(data, &mapModel)
	if err != nil {
		log.Printf("Error unmarshaling map model: %v\n", err)
		return
	}

	e.currentMap = &mapModel
	e.updateMap(true)
	e.overviewMenu.updateTree(e.currentMap)
}

func (e *Editor) updateMap(clearMemory bool) {
	e.rootMapNode.RemoveAll(clearMemory)
	e.nodeIndex = make(map[string]*renderer.Node)

	cd := countdown.Countdown{}
	cd.Start(countGeometries(e.currentMap.Root))
	e.openProgressBar()
	e.setProgressBar(0)
	e.setProgressTime("Loading Map...")

	updateProgress := func() {
		cd.Count()
		e.setProgressBar(cd.PercentageComplete() / 5)
		e.setProgressTime(fmt.Sprintf("Loading Map... %v seconds remaining", cd.SecondsRemaining()))
	}

	var updateNode func(srcModel *editorModels.NodeModel, destNode *renderer.Node)
	updateNode = func(srcModel *editorModels.NodeModel, destNode *renderer.Node) {
		if srcModel.Geometry != nil {
			geometry, err := assets.ImportObjCached(*srcModel.Geometry)
			if err == nil {
				destNode.Add(geometry)
			}
			updateProgress()
		}
		destNode.SetScale(srcModel.Scale)
		destNode.SetTranslation(srcModel.Translation)
		destNode.SetOrientation(srcModel.Orientation)
		for _, childModel := range srcModel.Children {
			newNode := renderer.CreateNode()
			destNode.Add(newNode)
			e.nodeIndex[childModel.Id] = newNode
			updateNode(childModel, newNode)
		}
	}

	updateNode(e.currentMap.Root, e.rootMapNode)
	e.closeProgressBar()
}

func countGeometries(nodeModel *editorModels.NodeModel) int {
	count := 0
	if nodeModel.Geometry != nil {
		count++
	}
	for _, child := range nodeModel.Children {
		count += countGeometries(child)
	}
	return count
}