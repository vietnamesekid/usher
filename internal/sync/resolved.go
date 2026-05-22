package sync

// Re-export types from the shared types package so callers can use
// either sync.ResolvedConfig or types.ResolvedConfig interchangeably.
import "github.com/vietnamesekid/usher/internal/types"

type ResolvedConfig = types.ResolvedConfig
type ResolvedMCPInstance = types.ResolvedMCPInstance
type ResolvedSkill = types.ResolvedSkill
