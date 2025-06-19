# yq-unity - Unity YAML Support for yq

This is a fork of [yq](https://github.com/mikefarah/yq) with added support for Unity's YAML asset files.

## Why yq-unity?

Unity uses a non-standard YAML format that includes:
- Custom tags like `!u!114` that break standard YAML parsers
- Multiple YAML documents in a single file connected by `fileID` references
- Complex nested structures spread across documents

Standard yq cannot parse these files correctly. yq-unity adds Unity-specific features to handle these formats.

## Installation

1. Clone the repository with the yq-unity submodule
2. Build yq-unity:
```bash
cd Tools/yq-unity
go build -o yq-unity yq.go
```

## Features

### 1. Unity YAML Format Support

The `-p unity` flag enables Unity YAML parsing:
```bash
yq-unity -p unity eval '.MonoBehaviour.m_Name' file.asset
```

### 2. Unity Extract Command

A specialized command for extracting data from Unity asset files:

```bash
# Extract all spawner nodes with distance properties
yq-unity unity-extract spawners file.asset

# List all nodes in the asset
yq-unity unity-extract nodes file.asset

# Extract specific property values
yq-unity unity-extract property --property Extents file.asset
```

### 3. Wrapper Script

A convenience wrapper script is provided at `Tools/yq-unity.sh`:
```bash
# Automatically uses Unity format for .asset files
./yq-unity.sh eval '.MonoBehaviour.m_Name' file.asset
```

## Usage Examples

### Finding Spawner Distances

Extract all spawner nodes with their distance-related properties:
```bash
yq-unity unity-extract spawners BackgroundObjects.asset
```

Output:
```
Pose Set Spawner:
  Extents: 150

Grid Spawner:
  Extents: 150
  GridCellSize: 5

Total spawners found: 2
```

### Listing All Nodes

Get a list of all nodes in an asset file:
```bash
yq-unity unity-extract nodes RocksAndSpawnArea.asset
```

### Analyzing Property Values

Find all unique values of a specific property:
```bash
yq-unity unity-extract property --property GridCellSize *.asset
```

Output:
```
Property 'GridCellSize' values:
  2: 1 occurrences
  4: 1 occurrences
  5: 1 occurrences
  7.5: 1 occurrences
  10: 1 occurrences
```

## Unity-Specific Operators

Two experimental operators are available (may need refinement):

- `unity_nodes` - Extract all nodes from Unity YAML
- `unity_spawners` - Extract spawner-specific data

These can be used with standard yq syntax:
```bash
yq-unity -p unity 'unity_nodes | keys' file.asset
```

## Common Unity Asset Properties

### Spawner Properties
- **Extents**: The spawning area radius (e.g., 150m)
- **GridCellSize**: Grid spacing for grid-based spawners (e.g., 5m, 10m)
- **MinimumDistance**: Minimum distance between spawned objects
- **RoadPoseDistance**: Distance along road for pose-based spawning
- **Spacing**: General spacing parameter

### Node Types
- **Grid Spawner**: Uses grid-based spawning with configurable cell size
- **Pose Set Spawner**: Uses predefined poses for object placement
- **Spawner Portal**: Connection point between spawner systems

## Limitations

1. The Unity YAML parser removes some Unity-specific data (like fileID references) to make the YAML parseable
2. Complex serialized data in `serializedData` fields may not be fully accessible
3. Cross-document references are not automatically resolved

## Technical Details

### How It Works

1. **Tag Removal**: Unity tags like `!u!114` are stripped during preprocessing
2. **Reference Handling**: `{fileID: ...}` references are converted to null values
3. **Document Separation**: Multiple documents are processed independently

### File Structure Support

Unity asset files typically contain:
- Multiple YAML documents (separated by `---`)
- Each document represents a Unity object with a unique fileID
- Objects reference each other via fileID values
- Complex data may be stored in serialized binary format

## Future Improvements

Potential enhancements could include:
- Automatic fileID reference resolution
- Support for serializedData binary parsing
- Graph visualization of node connections
- Direct Unity project integration

## Contributing

When adding new features:
1. Test with various Unity asset file types
2. Ensure backward compatibility with standard yq features
3. Document Unity-specific behaviors

## License

This fork maintains the same license as the original yq project.