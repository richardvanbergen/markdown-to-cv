package mcp

import (
	"github.com/mark3labs/mcp-go/server"
)

// Server wraps the MCP server with tools for interactive CV optimization.
type Server struct {
	mcpServer *server.MCPServer
}

// NewServer creates a new MCP server configured with the write_optimized_resume tool.
func NewServer(ctx *InteractiveContext) *Server {
	mcpServer := server.NewMCPServer(
		"m2cv",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Register the write_optimized_resume tool
	tool := NewWriteOptimizedResumeTool()
	handler := WriteOptimizedResumeHandler(ctx.ApplicationDir)
	mcpServer.AddTool(tool, handler)

	return &Server{
		mcpServer: mcpServer,
	}
}

// Serve starts the MCP server on stdio.
func (s *Server) Serve() error {
	return server.ServeStdio(s.mcpServer)
}
