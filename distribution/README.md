# Paquetes para los directorios

- `claude/`: enviar esta ruta como plugin path del repositorio `rapibase/skills`. Contiene los cuatro skills, el MCP remoto con OAuth y el icono PNG. La validación local de `claude plugin validate` pasa; el portal realiza análisis adicionales.
- `openai/`: contenido del ZIP para el directorio de OpenAI. El MCP está declarado desde la primera versión mediante `mcp.json` y OAuth con registro dinámico.

Los skills de ambos paquetes son copias de `../skills/`; al actualizar los originales, sincronizar ambas carpetas antes de publicar. No comprimir `distribution/` completo: la raíz de cada paquete es `claude/` u `openai/`.

## Antes del envío

- Confirmar en el portal de Claude los detalles del bloqueo de credenciales y las cuatro retenciones de política. Se retiró de la guía el ejemplo que enviaba `RAPIBASE_TOKEN` como encabezado del MCP, pero no tenemos el archivo y la línea exactos del hallazgo original.
- Confirmar una URL pública de documentación y una página HTTPS de soporte. Los manifiestos usan `https://rapibase.com/docs` y `https://rapibase.com/support`; estas rutas aún no se han confirmado como publicadas.
- Revisar los colores de marca y preparar, si hace falta, variantes oscuras del icono. Las cuatro imágenes de OpenAI son copias del PNG fuente de 512 px.
- Crear la cuenta de revisión, los casos de uso y el video pedidos por OpenAI, y completar la verificación de dominio en el control plane.
- Validar en el portal de cada empresa después de subir el commit revisado. La validación local de Claude comprueba menos reglas que el portal.
