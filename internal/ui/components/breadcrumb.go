package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/miladbeigi/vaultui/internal/ui/styles"
)

// Breadcrumb renders a styled breadcrumb trail from path segments.
// The last segment is highlighted as "active". The optional suffix
// is appended as the final active segment (e.g. "versions").
func Breadcrumb(mount, path, suffix string, width int) string {
	sep := styles.BreadcrumbStyle.Render(" ▸ ")

	var segments []string
	segments = append(segments, mount)

	if path != "" {
		for _, seg := range strings.Split(strings.TrimSuffix(path, "/"), "/") {
			if seg != "" {
				segments = append(segments, seg)
			}
		}
	}
	if suffix != "" {
		segments = append(segments, suffix)
	}

	parts := make([]string, len(segments))
	for i, seg := range segments {
		if i == len(segments)-1 {
			parts[i] = styles.BreadcrumbActiveStyle.Render(seg)
		} else {
			parts[i] = styles.SubtleStyle.Render(seg)
		}
	}

	crumb := strings.Join(parts, sep)
	return lipgloss.NewStyle().Width(width).PaddingBottom(1).Render(crumb)
}
