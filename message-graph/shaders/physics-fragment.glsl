#version 300 es
precision highp float;

// This fragment shader is minimal since we're using transform feedback
// for the physics simulation. The actual computation happens in the vertex shader.
// This shader just outputs a dummy color since we're not rendering to screen.

out vec4 fragColor;

void main() {
    // Output a dummy color - this won't be visible since we're using transform feedback
    fragColor = vec4(0.0, 0.0, 0.0, 0.0);
}

