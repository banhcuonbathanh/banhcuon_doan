# Create feature branch

git checkout -b feature-account-interface-service

# Commit changes

git add .
git commit -m "Add feature X"

# Push to GitHub (first time)

git push -u origin feature-x

# Later updates (after commits)

git push

# Delete branch after merge (local & remote)

git branch -d feature-x
git push origin --delete feature-x

git push --set-upstream origin feature-account

# Ensure you're on master and up-to-date

git checkout master
git pull origin master

# Merge feature branch

git merge feature-account

# If conflicts occur, resolve them then:

git add .
git commit -m "Merge feature-account with conflict resolution"

# Push to GitHub

git push origin master

# Cleanup (optional)

git branch -d feature-account
git push origin --delete feature-account