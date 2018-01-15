package galaxy

import (
	"testing"

	"github.com/stvp/assert"
)

func TestMaterials_category(t *testing.T) {
	galaxy, err := OpenGalaxy("test/systems.json", "../bcplus.d/data")
	assert.Nil(t, err, "open galaxy", err)
	matCat := galaxy.MatCategory("phosphorus")
	assert.Equal(t, Raw, matCat)
}
