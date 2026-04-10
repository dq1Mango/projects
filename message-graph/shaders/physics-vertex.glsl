#version 300 es

// Input attributes for current node state
in vec2 a_position;     // Current position (x, y)
in vec2 a_velocity;     // Current velocity (vx, vy)
in float a_mass;        // Node mass (affects forces)
in float a_radius;      // Node radius
in float a_fixed;       // 1.0 if node is fixed (being dragged), 0.0 otherwise

// Uniforms for simulation parameters
uniform float u_repulsionForce;   // Global repulsion strength
uniform float u_springConstant;   // Spring constant for edges
uniform float u_linkDistance;     // Target distance for springs
uniform float u_damping;          // Velocity damping factor
uniform float u_deltaTime;        // Time step
uniform vec2 u_canvasSize;        // Canvas dimensions for boundary checking

// Uniforms for node data (all nodes)
uniform sampler2D u_nodePositions;  // Texture containing all node positions
uniform sampler2D u_nodeData;       // Texture containing mass, radius, fixed flag
uniform int u_nodeCount;            // Total number of nodes

// Uniforms for edge data
uniform sampler2D u_edgeData;       // Texture containing edge information (from, to, weight)
uniform int u_edgeCount;            // Total number of edges

// Output attributes (for transform feedback)
out vec2 v_newPosition;
out vec2 v_newVelocity;

// Current node index (derived from gl_VertexID)
int nodeIndex;

// Helper function to get node data from texture
vec4 getNodePosition(int index) {
    int texWidth = textureSize(u_nodePositions, 0).x;
    int x = index % texWidth;
    int y = index / texWidth;
    return texelFetch(u_nodePositions, ivec2(x, y), 0);
}

vec4 getNodeData(int index) {
    int texWidth = textureSize(u_nodeData, 0).x;
    int x = index % texWidth;
    int y = index / texWidth;
    return texelFetch(u_nodeData, ivec2(x, y), 0);
}

vec4 getEdgeData(int index) {
    int texWidth = textureSize(u_edgeData, 0).x;
    int x = index % texWidth;
    int y = index / texWidth;
    return texelFetch(u_edgeData, ivec2(x, y), 0);
}

void main() {
    nodeIndex = gl_VertexID;
    
    // Initialize forces
    vec2 totalForce = vec2(0.0);
    
    // Skip force calculation if this node is fixed (being dragged)
    if (a_fixed > 0.5) {
        v_newPosition = a_position;
        v_newVelocity = vec2(0.0);
        return;
    }
    
    // Calculate repulsion forces from all other nodes
    for (int i = 0; i < u_nodeCount; i++) {
        if (i == nodeIndex) continue;
        
        vec4 otherPos = getNodePosition(i);
        vec2 delta = a_position - otherPos.xy;
        float distance = length(delta);
        
        // Avoid division by zero and self-interaction
        if (distance > 0.1) {
            // Coulomb-like repulsion: F = k * q1 * q2 / r^2
            float force = u_repulsionForce / (distance * distance);
            vec2 direction = normalize(delta);
            totalForce += direction * force;
        }
    }
    
    // Calculate spring forces from connected edges
    for (int i = 0; i < u_edgeCount; i++) {
        vec4 edge = getEdgeData(i);
        int fromNode = int(edge.x);
        int toNode = int(edge.y);
        float weight = edge.z;
        
        // Check if this node is part of this edge
        int otherNodeIndex = -1;
        if (fromNode == nodeIndex) {
            otherNodeIndex = toNode;
        } else if (toNode == nodeIndex) {
            otherNodeIndex = fromNode;
        }
        
        if (otherNodeIndex >= 0) {
            vec4 otherPos = getNodePosition(otherNodeIndex);
            vec2 delta = otherPos.xy - a_position;
            float distance = length(delta);
            
            if (distance > 0.1) {
                // Hooke's law: F = k * (distance - restLength)
                float targetDistance = u_linkDistance;
                float force = (distance - targetDistance) * u_springConstant * weight;
                vec2 direction = normalize(delta);
                totalForce += direction * force;
            }
        }
    }
    
    // Update velocity with forces and damping
    vec2 acceleration = totalForce / a_mass;
    vec2 newVelocity = (a_velocity + acceleration * u_deltaTime) * u_damping;
    
    // Update position
    vec2 newPosition = a_position + newVelocity * u_deltaTime;
    
    // Boundary collision detection
    float margin = a_radius;
    if (newPosition.x < margin) {
        newPosition.x = margin;
        newVelocity.x = -newVelocity.x * 0.5; // Bounce with energy loss
    }
    if (newPosition.x > u_canvasSize.x - margin) {
        newPosition.x = u_canvasSize.x - margin;
        newVelocity.x = -newVelocity.x * 0.5;
    }
    if (newPosition.y < margin) {
        newPosition.y = margin;
        newVelocity.y = -newVelocity.y * 0.5;
    }
    if (newPosition.y > u_canvasSize.y - margin) {
        newPosition.y = u_canvasSize.y - margin;
        newVelocity.y = -newVelocity.y * 0.5;
    }
    
    // Output new state
    v_newPosition = newPosition;
    v_newVelocity = newVelocity;
    
    // Set gl_Position for rendering (not used in physics simulation)
    gl_Position = vec4(0.0);
}
