package mesh

import (
	"reflect"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/go3mf/internal/meshinfo"
	"github.com/stretchr/testify/mock"
)

func TestNewMesh(t *testing.T) {
	tests := []struct {
		name string
		want *Mesh
	}{
		{"base", &Mesh{
			beamLattice:        *newbeamLattice(),
			informationHandler: *meshinfo.NewHandler(),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want.faceStructure.informationHandler = &tt.want.informationHandler
			if got := NewMesh(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMesh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMesh_Clone(t *testing.T) {
	tests := []struct {
		name    string
		m       *Mesh
		want    *Mesh
		wantErr bool
	}{
		{"base", NewMesh(), NewMesh(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.Clone()
			if (err != nil) != tt.wantErr {
				t.Errorf("Mesh.Clone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mesh.Clone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMesh_Clear(t *testing.T) {
	tests := []struct {
		name string
		m    *Mesh
	}{
		{"base", new(Mesh)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.Clear()
		})
	}
}

func TestMesh_InformationHandler(t *testing.T) {
	tests := []struct {
		name string
		m    *Mesh
		want meshinfo.Handler
	}{
		{"created", &Mesh{informationHandler: *meshinfo.NewHandler()}, *meshinfo.NewHandler()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := *tt.m.InformationHandler(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mesh.InformationHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMesh_ClearInformationHandler(t *testing.T) {
	tests := []struct {
		name string
		m    *Mesh
	}{
		{"base", &Mesh{informationHandler: *meshinfo.NewHandler()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.informationHandler.AddBaseMaterialInfo(0)
			tt.m.informationHandler.AddNodeColorInfo(0)
			tt.m.ClearInformationHandler()
			if tt.m.informationHandler.InformationCount() != 0 {
				t.Error("Mesh.ClearInformationHandler expected to clear the handler")
			}
		})
	}
}

func TestMesh_Merge(t *testing.T) {
	type args struct {
		mesh   *MockMergeableMesh
		matrix mgl32.Mat4
	}
	tests := []struct {
		name    string
		m       *Mesh
		args    args
		nodes   uint32
		faces   uint32
		wantErr bool
	}{
		{"error1", new(Mesh), args{new(MockMergeableMesh), mgl32.Ident4()}, 0, 0, false},
		{"error2", new(Mesh), args{new(MockMergeableMesh), mgl32.Ident4()}, 1, 0, false},
		{"base", new(Mesh), args{new(MockMergeableMesh), mgl32.Ident4()}, 1, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.mesh.On("InformationHandler").Return(meshinfo.NewHandler())
			tt.args.mesh.On("NodeCount").Return(tt.nodes)
			tt.args.mesh.On("Node", mock.Anything).Return(new(Node))
			tt.args.mesh.On("FaceCount").Return(tt.faces)
			tt.args.mesh.On("Face", mock.Anything).Return(new(Face))
			tt.args.mesh.On("BeamCount").Return(uint32(0))
			if err := tt.m.Merge(tt.args.mesh, tt.args.matrix); (err != nil) != tt.wantErr {
				t.Errorf("Mesh.Merge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMesh_CheckSanity(t *testing.T) {
	tests := []struct {
		name string
		m    *Mesh
		want bool
	}{
		{"new", NewMesh(), true},
		{"nodefail", &Mesh{nodeStructure: nodeStructure{maxNodeCount: 1, nodes: make([]Node, 2)}}, false},
		{"facefail", &Mesh{faceStructure: faceStructure{maxFaceCount: 1, faces: make([]Face, 2)}}, false},
		{"beamfail", &Mesh{beamLattice: beamLattice{maxBeamCount: 1, beams: make([]Beam, 2)}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.CheckSanity(); got != tt.want {
				t.Errorf("Mesh.CheckSanity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMesh_ApproxEqual(t *testing.T) {
	type args struct {
		mesh *Mesh
	}
	tests := []struct {
		name string
		m    *Mesh
		args args
		want bool
	}{
		{"base", NewMesh(), args{nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.ApproxEqual(tt.args.mesh); got != tt.want {
				t.Errorf("Mesh.ApproxEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMesh_StartCreation(t *testing.T) {
	type args struct {
		opts CreationOptions
	}
	tests := []struct {
		name string
		m    *Mesh
		args args
	}{
		{"default", NewMesh(), args{CreationOptions{CalculateConnectivity: false}}},
		{"connectivity", NewMesh(), args{CreationOptions{CalculateConnectivity: true}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.StartCreation(tt.args.opts)
			if tt.args.opts.CalculateConnectivity && tt.m.nodeStructure.vectorTree == nil {
				t.Error("Mesh.StartCreation() should have created the vector tree")
				return
			}
			if !tt.args.opts.CalculateConnectivity && tt.m.nodeStructure.vectorTree != nil {
				t.Error("Mesh.StartCreation() shouldn't have created the vector tree")
				return
			}
		})
	}
}

func TestMesh_EndCreation(t *testing.T) {
	tests := []struct {
		name string
		m    *Mesh
	}{
		{"base", NewMesh()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.StartCreation(CreationOptions{CalculateConnectivity: true})
			tt.m.EndCreation()
			if tt.m.nodeStructure.vectorTree != nil {
				t.Error("Mesh.StartCreation() should have deleted the vector tree")
			}
		})
	}
}

func TestMesh_FaceNodes(t *testing.T) {
	m := NewMesh()
	n1 := m.AddNode(mgl32.Vec3{0.0, 0.0, 0.0})
	n2 := m.AddNode(mgl32.Vec3{20.0, -20.0, 0.0})
	n3 := m.AddNode(mgl32.Vec3{0.0019989014, 0.0019989014, 0.0})
	m.AddFace(n1.Index, n2.Index, n3.Index)
	type args struct {
		i uint32
	}
	tests := []struct {
		name  string
		m     *Mesh
		args  args
		want  *Node
		want1 *Node
		want2 *Node
	}{
		{"base", m, args{0}, n1, n2, n3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := tt.m.FaceNodes(tt.args.i)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mesh.FaceNodes() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Mesh.FaceNodes() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("Mesh.FaceNodes() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestMesh_IsManifoldAndOriented(t *testing.T) {
	tests := []struct {
		name string
		m    *Mesh
		want bool
	}{
		{"valid", &Mesh{
			nodeStructure: nodeStructure{nodes: []Node{{Index: 0}, {Index: 1}, {Index: 2}, {Index: 3}}},
			faceStructure: faceStructure{faces: []Face{
				{NodeIndices: [3]uint32{0, 1, 2}},
				{NodeIndices: [3]uint32{0, 3, 1}},
				{NodeIndices: [3]uint32{0, 2, 3}},
				{NodeIndices: [3]uint32{1, 3, 2}},
			}},
		}, true},
		{"nonmanifold", &Mesh{
			nodeStructure: nodeStructure{nodes: []Node{{Index: 0}, {Index: 1}, {Index: 2}, {Index: 3}}},
			faceStructure: faceStructure{faces: []Face{
				{NodeIndices: [3]uint32{0, 1, 2}},
				{NodeIndices: [3]uint32{0, 1, 3}},
				{NodeIndices: [3]uint32{0, 2, 3}},
				{NodeIndices: [3]uint32{1, 2, 3}},
			}},
		}, false},
		{"empty", NewMesh(), false},
		{"2nodes", &Mesh{
			nodeStructure: nodeStructure{nodes: make([]Node, 2)},
			faceStructure: faceStructure{faces: make([]Face, 3)},
		}, false},
		{"2faces", &Mesh{
			nodeStructure: nodeStructure{nodes: make([]Node, 3)},
			faceStructure: faceStructure{faces: make([]Face, 2)},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.IsManifoldAndOriented(); got != tt.want {
				t.Errorf("Mesh.IsManifoldAndOriented() = %v, want %v", got, tt.want)
			}
		})
	}
}
