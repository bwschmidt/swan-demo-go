# Contributing

When contributing changes to this repository, please first create an issue on 
the repository before making the change. 

## Branching Strategies & Pull Requests

This repository uses mainline branching also knows as 
[GitHub flow](https://guides.github.com/introduction/flow/)

## Commit Messages

To ensure that everyone can clearly follow changes to the source code, commit 
comments need to be clear and detailed. 

Commits have two components.

- Subject – the 50 or so characters that appear in all history views.
- Description – more verbose text which explains the commit, how it’s been 
tested and links to any additional information.

### Subject 

The subject or single first line of the commit should include:

- its nature (is it a fix for a bug, a new feature, an optimization, ...)
- its importance, which generally reflects the risk of merging/not merging it
- what area it applies to (eg: device detection, examples, unit tests, code comments, ...)

It's important to make these three criteria easy to spot in the patch's subject, 
because it's the first (and sometimes the only) thing which is read when 
reviewing patches to find which ones need to be backported to older versions.

Specifically, bugs must be clearly easy to spot so that they're never missed. 
Any patch fixing a bug must have the "BUG" tag in its subject. Most common patch 
types include:

|Tag|Description|
|---|---|
|BUG|Fix for a bug. The severity of the bug should also be indicated when known.|
|CLEANUP|Code cleanup, silence of warnings, etc... theoretically no impact. By nature, a clean-up is always minor.|
|REORG|Code reorganization. Some blocks may be moved to other places, some important checks might be swapped, etc... These changes always present a risk of regression. For this reason, they should never be mixed with any bug fix nor functional change. Code is only moved as-is. Indicating the risk of breakage is highly recommended.|
|BUILD|Updates or fixes for build issues. Changes to makefiles, POMs, csproj, etc… also fall into this category. The risk of breakage should be indicated if known. It is also appreciated to indicate what platforms and/or configurations were tested after the change.|
|OPTIM|Some code was optimised. Sometimes if the regression risk is very low and the gains significant, such patches may be merged in the stable branch. Depending on the amount of code changed or replaced and the level of trust the author has in the change, the risk of regression should be indicated.|
|FEAT|A new feature has been added. The commit should provide a summary of the feature and details of how to find out more about the feature. Such details may appear in auto generated documentation such as Javadocs, Sandcastle, etc… The comments should not attempt to duplicate the developer documentation.|
|DOC|Documentation updates or fixes. No code is affected, no need to upgrade. These patches can also be sent right after a new feature, to document it where the documentation is part of the repository via Javadocs, Sandcastle, etc…|
|EXAMPLE|Changes to example files.|
|TESTS|Regression test or other test files. No code is affected, no need to upgrade.|
|CONT|Change to published content, e.g webpage, documentation|
|CONF|Change to tracked configuration files, e.g. appsettings.json...|

All commits must be tagged. Additionally, the importance of the patch should be 
indicated when known. A single upper-case word is preferred.

|Tag|Description|
|---|---|
|MINOR|Minor change, very low risk of impact. For a bug, it generally indicates an annoyance, nothing more.|
|MEDIUM|Medium risk, may cause unexpected regressions of low importance or which may quickly be discovered. For a bug, it generally indicates something odd which requires changing the configuration in an undesired way to work around the issue.|
|MAJOR|Major risk of hidden regression. This happens when I rearrange large parts of code, when I play with timeouts, with variable initializations, etc... For a bug, it indicates severe reliability issues for which workarounds are identified with or without performance impacts.|
|CRITICAL|Medium-term reliability or security is at risk and workarounds, if they exist, might not always be acceptable. An upgrade is absolutely required. A maintenance release may be emitted even if only one of these bugs are fixed. Note that this tag is only used with bugs. Such patches must indicate what is the first version affected, and if known, the commit ID which introduced the issue.|

It is desired that AT LEAST one of the two criteria tags is reported in the 
commit subject. If both are present, then they should be delimited with a slash 
('/'). They should always appear before a colon (‘:’).Thus, all of the following 
subjects are valid:

Examples of subjects:

- DOC: Document options for store implementation in go
- DOC/MAJOR: Reorganize the whole document and change indenting
- BUG/MEDIUM: Modified the OWID to place the first field as the version.
- BUG/CRITICAL: Preferences page handles missing data as a Bad Request HTTP 
status code.
- BUG/MINOR: The cache size was being halved resulting in a smaller cache than 
expected
- FEAT/MEDIUM: Added access control to SWAN.
- BUILD: Changed the Makefile...
- BUILD/MEDIUM: Added build support for Clang in OSx.
- OPTIM/MINOR: Changed use of Slices to Arrays
- REORG/MEDIUM: Moved the go classes into their own package

### Description

New features included in a release may be summarised in the README.MD or other 
high level information. Such summaries should not duplicate the commit 
description.

The description in the commit should provide more information and be the only 
place that changes are described in detail to avoid duplication.

The description should explain the following:

- Why. This could optionally reference an issue or URL explaining more about the 
background. Such links could be forum posts, or public Git issues. They should 
not reference private internal issues not publicly visible.
- What. A summary of what has changed which does not duplicate new code 
comments, or information that can clearly been seem when comparing the previous 
version to the committed version.
- How tested. If appropriate how the change has been tested.
- Warnings and advice. Guidance to the developer taking the commit about how to 
use it or test it.
