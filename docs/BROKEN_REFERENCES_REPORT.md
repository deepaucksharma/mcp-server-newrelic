# Broken References and Outdated File Paths Report

## Summary

After searching through all markdown files in the repository, I found several broken references and outdated file paths that need to be fixed. Most references point to files that were either deleted or moved to new locations.

## Broken References Found

### 1. QUICKSTART.md References
- **Location**: `docs/architecture/documentation-structure.md` (line 49)
- **Location**: `docs/architecture/ecosystem-overview.md` (line 291)
- **Issue**: References `QUICKSTART.md` which doesn't exist
- **Fix**: Should reference quick start section in main README.md or create a new quickstart guide in docs/guides/

### 2. TROUBLESHOOTING.md References
- **Location**: `docs/README.md` (already updated to `./guides/troubleshooting.md`)
- **Location**: `docs/architecture/documentation-structure.md` (line 30)
- **Issue**: Still references old path
- **Fix**: Update to `guides/troubleshooting.md`

### 3. CONTRIBUTING.md References
- **Location**: `docs/architecture/documentation-structure.md` (line 80)
- **Location**: `cmd/uds/README.md` (line 293)
- **Issue**: References `CONTRIBUTING.md` and `guides/contributing.md`
- **Fix**: Should reference development guide or create a contributing guide

### 4. API_REFERENCE_V2.md References
- **Location**: `docs/guides/refactoring.md` (line 316) - **FIXED**
- **Location**: `docs/guides/migration.md` (line 408) - **FIXED**
- **Issue**: References deleted file `API_REFERENCE_V2.md`
- **Fix**: Updated to `../api/reference.md`

### 5. Python Implementation References
- **Location**: `docs/technical/platform-spec.md` (lines 7, 179, 233)
- **Location**: `docs/architecture/complete-reference.md` (lines 96-99, 756-787)
- **Location**: `ROADMAP_2025.md` (line 233)
- **Issue**: Still mentions Python implementation and dual implementation
- **Fix**: Update to reflect Go-only implementation or clarify deprecation status

### 6. Other Outdated Documentation References
In `docs/architecture/documentation-structure.md`:
- Line 30: References many files in docs/ root that are now in subdirectories
- Line 50: `DISCOVERY_ECOSYSTEM_OVERVIEW.md` - doesn't exist
- Line 55: `ARCHITECTURE_COMPLETE.md` - should be `architecture/complete-reference.md`
- Line 59: `REFACTORING_GUIDE.md` - should be `guides/refactoring.md`
- Line 60: `TOOL_GRANULARITY_ENHANCEMENT.md` - doesn't exist
- Line 64: `philosophy/DISCOVERY_PHILOSOPHY.md` - doesn't exist
- Line 70: `WORKFLOW_PATTERNS_GUIDE.md` - doesn't exist
- Line 75: `ux/INTERACTIVE_LEARNING_GUIDE.md` - doesn't exist
- Line 80: `DEVELOPMENT.md` - should be `guides/development.md`

In `docs/architecture/ecosystem-overview.md`:
- Line 21: `philosophy/DISCOVERY_PHILOSOPHY.md` - doesn't exist
- Line 30: `TECHNICAL_PLATFORM_SPEC.md` - should be `technical/platform-spec.md`
- Line 34: `ARCHITECTURE_COMPLETE.md` - should be `architecture/complete-reference.md`
- Line 60: `TOOL_GRANULARITY_ENHANCEMENT.md` - doesn't exist
- Line 66: `WORKFLOW_PATTERNS_GUIDE.md` - doesn't exist
- Line 70: `WORKFLOW_PATTERNS_GUIDE.md` - doesn't exist (duplicate)
- Line 76: `DATA_OBSERVABILITY_TOOLKIT.md` - doesn't exist
- Line 86: `REFACTORING_GUIDE.md` - should be `guides/refactoring.md`
- Line 87: `MIGRATION_GUIDE.md` - should be `guides/migration.md`
- Line 88: `DEVELOPMENT.md` - should be `guides/development.md`
- Line 292: `docs/QUICKSTART.md` - doesn't exist
- Line 293: `docs/DISCOVERY_PHILOSOPHY.md` - doesn't exist
- Line 294: `docs/INTERACTIVE_DISCOVERY_EXPERIENCES.md` - doesn't exist
- Line 295: `docs/WOW_EXPERIENCE_SUMMARY.md` - doesn't exist

In `docs/guides/refactoring.md`:
- Line 315: `./DISCOVERY_FIRST_ARCHITECTURE.md` - should be `../architecture/discovery-first.md`
- Line 316: `./DISCOVERY_DRIVEN_INVESTIGATION_EXAMPLES.md` - should be `../examples/DISCOVERY_DRIVEN_INVESTIGATION_EXAMPLES.md`
- Line 317: `./WORKFLOW_PATTERNS_GUIDE.md` - doesn't exist
- Line 318: `./API_REFERENCE_V2.md` - should be `../api/reference.md`

In `docs/guides/migration.md`:
- Line 407: `./DISCOVERY_FIRST_ARCHITECTURE.md` - should be `../architecture/discovery-first.md`
- Line 409: `./WORKFLOW_PATTERNS_GUIDE.md` - doesn't exist

In `README.md`:
- Line 196: References `./docs/DEVELOPMENT.md` - should be `./docs/guides/development.md`
- Line 228: References `./docs/DEVELOPMENT.md` - should be `./docs/guides/development.md`

In `CLAUDE.md`:
- Line 196: References `./docs/DEVELOPMENT.md` - should be `./docs/guides/development.md`

## Recommendations

1. **Create Missing Guides**:
   - `docs/guides/quickstart.md` - A dedicated quick start guide
   - `docs/guides/contributing.md` - Contributing guidelines
   - `docs/examples/workflow-patterns.md` - Workflow patterns guide

2. **Update All References**:
   - Replace `API_REFERENCE_V2.md` with `api/reference.md` ✓ COMPLETED
   - Update all documentation paths to reflect new structure
   - Remove or update Python implementation references

3. **Consolidate Philosophy Documents**:
   - Create `docs/philosophy/discovery-philosophy.md` if needed
   - Or update references to point to existing discovery-first documentation

4. **Fix Documentation Structure File**:
   - `docs/architecture/documentation-structure.md` needs major updates
   - Many referenced files don't exist or have moved

5. **Update Ecosystem Overview**:
   - `docs/architecture/ecosystem-overview.md` has many broken links
   - Needs comprehensive update to reflect current structure

## Files Needing Immediate Attention

1. `docs/architecture/documentation-structure.md` - Most broken references
2. `docs/architecture/ecosystem-overview.md` - Second most broken references
3. `docs/guides/refactoring.md` - API reference needs updating
4. `docs/guides/migration.md` - API reference needs updating
5. `README.md` - Development guide references
6. `CLAUDE.md` - Development guide reference