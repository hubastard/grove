#version 330 core
in vec4 vColor;
in vec2 vUV;
in float vTexIndex;

out vec4 FragColor;

// Bind up to 16 texture slots as an array
uniform sampler2D uTex[16];

void main() {
    int idx = int(vTexIndex);
    vec4 tex = texture(uTex[idx], vUV);
    FragColor = tex * vColor;
}
