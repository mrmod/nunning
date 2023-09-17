Each file in here is a BuildTool.

# BuildTool

A BuildTool takes Inputs and creates Outputs. They have Options (BuildToolOptions). Among their options are the Arguments to pass to the BuildTool's Executor.

A BuildTool's Executor (eg: `gcc`, `go`) is what generates the Build Outputs.

# Installing a BuildTools

BuildTools are installed by convention. 
* In the path `./buildtools`
* Place a Dockerfile name `Dockerfile.$BuildToolName`

# BuildTool Interface

A BuildTool must accept Inputs and produce Outputs. It must accept Arguments to apply to the transformation of Inputs to Outputs.

## Example: Installing a Go Build Tool

```
cp Dockerfile buildtools/Dockerfile.go
```