#version 330 core
layout(location=0) in vec2 aPos;
layout(location=1) in vec4 aColor;
layout(location=2) in vec2 aUV;
layout(location=3) in float aTexIndex;

uniform mat4 uVP;

out vec4 vColor;
out vec2 vUV;
out float vTexIndex;

void main() {
    vColor = aColor;
    vUV = aUV;
    vTexIndex = aTexIndex;
    gl_Position = uVP * vec4(aPos, 0.0, 1.0);
}
