# Contribution Guidelines

The Cloud Foundry team uses GitHub and accepts contributions via
[pull request](https://help.github.com/articles/using-pull-requests).

## Contributor License Agreement

Follow these steps to make a contribution to any of our open source repositories:

1. Ensure that you have completed our CLA Agreement for
  [individuals](http://www.cloudfoundry.org/individualcontribution.pdf) or
  [corporations](http://www.cloudfoundry.org/corpcontribution.pdf).

1. Set your name and email (these should match the information on your submitted CLA)

        git config --global user.name "Firstname Lastname"
        git config --global user.email "your_email@example.com"

## General Workflow
1. Fork the repository
1. Create a branch (`git checkout -b my_feature`)
1. Make changes on your branch
1. [Run the tests](https://github.com/cf-guardian/guardian#Testing)
1. Push to your fork (`git push origin my_feature`) and submit a pull request

We prefer pull requests with very small, single commits with a single purpose.

Your pull request is much more likely to be accepted if it:
* Includes tests
* Is small and focused
* Conforms to standard Go formatting conventions (`go fmt`)
* Contains a message explaining the intent of your change.
