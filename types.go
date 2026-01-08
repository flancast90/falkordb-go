package falkordb

import (
	"fmt"
	"strings"
	"time"
)

// ConstraintType represents the type of constraint.
type ConstraintType string

const (
	// ConstraintMandatory requires the property to be present.
	ConstraintMandatory ConstraintType = "MANDATORY"

	// ConstraintUnique requires the property value to be unique.
	ConstraintUnique ConstraintType = "UNIQUE"
)

// EntityType represents the type of graph entity.
type EntityType string

const (
	// EntityNode represents a graph node.
	EntityNode EntityType = "NODE"

	// EntityRelationship represents a graph relationship/edge.
	EntityRelationship EntityType = "RELATIONSHIP"
)

// Node represents a graph node with labels and properties.
type Node struct {
	// ID is the internal node identifier.
	ID int64

	// Labels are the node's labels.
	Labels []string

	// Properties are the node's key-value properties.
	Properties map[string]interface{}
}

// String returns a string representation of the node.
func (n *Node) String() string {
	labels := strings.Join(n.Labels, ":")
	return fmt.Sprintf("(:%s %v)", labels, n.Properties)
}

// Edge represents a graph relationship/edge.
type Edge struct {
	// ID is the internal edge identifier.
	ID int64

	// RelationshipType is the edge's type.
	RelationshipType string

	// SourceID is the ID of the source node.
	SourceID int64

	// DestinationID is the ID of the destination node.
	DestinationID int64

	// Properties are the edge's key-value properties.
	Properties map[string]interface{}
}

// String returns a string representation of the edge.
func (e *Edge) String() string {
	return fmt.Sprintf("-[:%s %v]->", e.RelationshipType, e.Properties)
}

// Path represents a sequence of nodes connected by edges.
type Path struct {
	// Nodes are the nodes in the path.
	Nodes []*Node

	// Edges are the edges connecting the nodes.
	Edges []*Edge
}

// Length returns the number of edges in the path.
func (p *Path) Length() int {
	return len(p.Edges)
}

// String returns a string representation of the path.
func (p *Path) String() string {
	if len(p.Nodes) == 0 {
		return "(empty path)"
	}

	var parts []string
	for i, node := range p.Nodes {
		parts = append(parts, node.String())
		if i < len(p.Edges) {
			parts = append(parts, p.Edges[i].String())
		}
	}
	return strings.Join(parts, "")
}

// Point represents a geographic point with latitude and longitude.
type Point struct {
	Latitude  float64
	Longitude float64
}

// String returns a string representation of the point.
func (p *Point) String() string {
	return fmt.Sprintf("POINT(%f %f)", p.Latitude, p.Longitude)
}

// Duration represents a temporal duration.
type Duration struct {
	Years        int
	Months       int
	Days         int
	Hours        int
	Minutes      int
	Seconds      int
	Nanoseconds  int
}

// ToDuration converts to a standard time.Duration.
// Note: Years and Months are approximated as 365 days and 30 days respectively.
func (d *Duration) ToDuration() time.Duration {
	total := time.Duration(d.Nanoseconds) * time.Nanosecond
	total += time.Duration(d.Seconds) * time.Second
	total += time.Duration(d.Minutes) * time.Minute
	total += time.Duration(d.Hours) * time.Hour
	total += time.Duration(d.Days) * 24 * time.Hour
	total += time.Duration(d.Months) * 30 * 24 * time.Hour
	total += time.Duration(d.Years) * 365 * 24 * time.Hour
	return total
}

// String returns the ISO 8601 duration string.
func (d *Duration) String() string {
	var parts []string

	if d.Years > 0 {
		parts = append(parts, fmt.Sprintf("%dY", d.Years))
	}
	if d.Months > 0 {
		parts = append(parts, fmt.Sprintf("%dM", d.Months))
	}
	if d.Days > 0 {
		parts = append(parts, fmt.Sprintf("%dD", d.Days))
	}

	if d.Hours > 0 || d.Minutes > 0 || d.Seconds > 0 {
		parts = append(parts, "T")
		if d.Hours > 0 {
			parts = append(parts, fmt.Sprintf("%dH", d.Hours))
		}
		if d.Minutes > 0 {
			parts = append(parts, fmt.Sprintf("%dM", d.Minutes))
		}
		if d.Seconds > 0 {
			parts = append(parts, fmt.Sprintf("%dS", d.Seconds))
		}
	}

	if len(parts) == 0 {
		return "PT0S"
	}
	return "P" + strings.Join(parts, "")
}

// DateTime represents a date and time value.
type DateTime struct {
	time.Time
}

// Date represents a date without time.
type Date struct {
	Year  int
	Month int
	Day   int
}

// ToTime converts to a time.Time at midnight UTC.
func (d *Date) ToTime() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
}

// String returns the ISO 8601 date string.
func (d *Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// Time represents a time without date.
type Time struct {
	Hour       int
	Minute     int
	Second     int
	Nanosecond int
}

// String returns the ISO 8601 time string.
func (t *Time) String() string {
	if t.Nanosecond > 0 {
		return fmt.Sprintf("%02d:%02d:%02d.%09d", t.Hour, t.Minute, t.Second, t.Nanosecond)
	}
	return fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second)
}
