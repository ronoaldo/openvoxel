#version 330 core
out vec4 FragColor;

in vec3 ourColor;
in vec2 TexCoord;

uniform sampler2D texture0;

void main() {
    //FragColor = vec4(ourColor, 1.0);
    //FragColor = vec4(TexCoord, TexCoord);
    FragColor = texture(texture0, TexCoord);
}