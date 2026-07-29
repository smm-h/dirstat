# dirstat improvements

## Text extension gaps

`.gd` (GDScript) and `.gdshader` (Godot shader language) are plaintext source code files but dirstat classifies them as non-text. The `file` command correctly identifies them as "ASCII text." They need to be added to `internal/config/data/text_extensions.txt`.

Other potentially missing extensions worth auditing: `.svx` (Svelte markdown), `.astro` (Astro), `.prisma` (Prisma schema), `.gql` (GraphQL alias), `.odin`, `.nim`, `.v` (Vlang), `.wgsl` (WebGPU shader), `.glsl`/`.hlsl`/`.frag`/`.vert` (shader languages).

## Audit text_extensions.txt completeness

The current list covers mainstream languages well. A systematic audit against popular language/framework extension lists would catch gaps. Game engine formats (Godot, Unity, Unreal) and shader languages are the most likely category of misses.

## The --exclude flag replaces defaults

When a user passes `--exclude`, it replaces the entire `defaultExcludes` list rather than adding to it. Users who want to exclude one additional directory must re-specify all 20+ defaults. Consider:
- An `--exclude-add` flag that appends to the defaults
- Or making `--exclude` additive, with `--exclude-only` for full replacement
