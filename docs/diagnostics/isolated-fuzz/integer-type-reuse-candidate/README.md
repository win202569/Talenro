# Integer-local type reuse candidate — NOT RUN

This isolated, unverified copy specializes only Assert-SupervisionPolicyInteger.
After the existing exact int/long gate, it performs the SDK extended-member and
public-descriptor inspections directly before bounds, avoiding a second type
query. No caller-supplied KnownType flag or input-acceptance cache is introduced.
Standalone Passive,Token,Bool,Shape and evaluator bodies remain unchanged.

The neighboring direct suite is byte-identical to the baseline suite. Original
product/tests and the deliberately defective negative control are untouched.
Do not confuse this candidate with that guard-loss control: ALL corresponding
passive checks must remain here. Keep the specialized block contract-equivalent
to Passive in future changes.

Prepared only: no candidate host or test has run; syntax/static checks are not
semantic or performance acceptance. No original replacement, full regression,
CodeProperty runtime coverage,review approval or timeout fix is implied.
Execute only after a separately approved pinned ordinary-host allowance;do not
reuse any spent/frozen budget. Git restores source/status,not run authority.
