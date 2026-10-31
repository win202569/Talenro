# Preparation-only revision V2

This leaf implements the approved preparation-only revision. The historical
driver/capture and failed packet stay unchanged. No runtime action, arbitrary
path, command, module import or process launcher is exposed.

Missing optional Role becomes null; existing string/null Role is preserved.
Mandatory path and SHA remain required. Regression evaluates only the actual
converter function AST, never the preparation body or historical driver.

Prepare validates explicit source mappings, not the old eleven-file leaf
count. Portable recovery files are excluded from source inputs. New outputs
use CreateNew; any partial preparation blocks another attempt.

Scope: at most two in-tool regression calls and one preparation invocation;
zero capture hosts/children/reviewers. Old allowances stay frozen. A manifest
does not grant runtime permission or restore a Session/Ack. Read the approved
plan and latest execution/recovery report before any action.

portable-summary.json retains the earlier stopped ledger. Later completion
data is separately named portable-completion.json; neither authorizes replay.
