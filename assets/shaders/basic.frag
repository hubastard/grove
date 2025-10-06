#version 330 core

in vec3 vColor;
in vec2 vUV;

uniform sampler2D uTex0;

out vec4 FragColor;

void main(){
    vec4 tex = texture(uTex0, vUV);
    FragColor = tex * vec4(vColor, 1.0);
}
