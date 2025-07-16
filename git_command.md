# Create feature branch
git checkout -b feature-account

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