# Native Go client instead of Azure CLI wrapper

We will build `pimctl` as a native Go client against Azure and Microsoft APIs instead of shelling out to `az`, because we want reliable UX, typed errors, testability, and high-quality CLI behavior. Azure CLI credentials may be reused where supported, but the CLI will not depend on invoking `az`.
