## Notes 

The static pool of repos works ok. Actually doing these things live takes time. 

Migration wizard takes time to run all the runs.

Codesmith needs ~10+ workflow runs before it is ok making a recommendation

Manually triggered a bunch of workflow runs, got codesmith to make the following recommendations:

- Integration Tests / Backend integration: 4vCPU -> 8vCPU
- Docker / Build backend image: 4vCPU -> 8vCPU
- CI / Backend unit tests: 4vCPU -> 8vCPU

Low Confidence recs:
- CI / Frontend: 4vCPU -> 2 vCPU
- CI / Frontend: 4vCPU -> 2 vCPU
- CI / Frontend: 4vCPU -> 2 vCPU

It reported the same test 3 times? Must be the matrix


## Unknowns

Is there a good way to track the evolution of the improvements as we onboard onto the platform?
    Maybe codesmith can tell me?

Could codesmith guide me through onboarding new features?

Explictly setting cache false feels fake, let's take that out?



