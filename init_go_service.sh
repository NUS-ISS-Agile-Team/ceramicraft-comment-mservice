#!/bin/bash
set -e
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <micro_short_name>"
    exit 1
fi
mic_short_name=$1
target_dir="../ceramicraft-${mic_short_name}-mservice/"
echo $target_dir
cp -r ./common "$target_dir"
echo "copy common to $target_dir"
cp -r client $target_dir
echo "copy client to $target_dir"
cp -r .github $target_dir
echo "copy .github to $target_dir"
cp -r server $target_dir
echo "copy server to $target_dir"
cp .gitignore $target_dir
echo "copy .gitignore to $target_dir"
cp sonar-project.properties $target_dir
echo "copy sonar-project.properties to $target_dir"

cd $target_dir
find . -name "*.go" -type f -exec sed -i '' "s/ceramicraft-mservice-template/ceramicraft-${mic_short_name}-mservice/g" {} +
find . -name "go.mod" -type f -exec sed -i '' "s/ceramicraft-mservice-template/ceramicraft-${mic_short_name}-mservice/g" {} +
find . -name "config.yml" -type f -exec sed -i '' "s/ceramicraft-mservice-template/ceramicraft-${mic_short_name}-mservice/g" {} +
find . -name "*.yml" -type f -exec sed -i '' "s/template_db/${mic_short_name}_db/g" {} +
find . -name "sonar-project.properties" -type f -exec sed -i '' "s/ceramicraft-mservice-template/ceramicraft-${mic_short_name}-mservice/g" {} +
find . -name "*.go" -type f -exec sed -i '' "s/template-ms/${mic_short_name}-ms/g" {} +

cd .github/
find . -name "*.yml" -type f -exec sed -i '' "s/ceramicraft-mservice-template/ceramicraft-${mic_short_name}-mservice/g" {} +
find . -name "*.yml" -type f -exec sed -i '' "s/template-ms/${mic_short_name}-ms/g" {} +


cd ..
cd common
go mod tidy
git add .
git commit -m "feat: init common for ${mic_short_name} service"
commit_hash=$(git log | head -n 1 | awk -F' ' '{print $2}')
echo "Commit hash: $commit_hash"
git push
echo "##############Complete common modification################"

cd ../client
sed -i '' "/ceramicraft-${mic_short_name}-mservice\/common/d" go.mod
go get "github.com/NUS-ISS-Agile-Team/ceramicraft-${mic_short_name}-mservice/common@$commit_hash"
go mod tidy
echo "#############Complete client modification################"

cd ../server
sed -i '' "/ceramicraft-${mic_short_name}-mservice\/common/d" go.mod
go get "github.com/NUS-ISS-Agile-Team/ceramicraft-${mic_short_name}-mservice/common@$commit_hash"
go mod tidy
swag init
echo "#############Complete server modification################"

cd ..
git add .
git commit -m "feat: init ceramicraft-${mic_short_name}-mservice"
git push
echo "##############Complete all modification for ceramicraft-${mic_short_name}-mservice################"
