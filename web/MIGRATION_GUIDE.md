# Ant Design to Tailwind + shadcn/ui Migration Guide

## Overview

This guide provides a comprehensive approach for migrating the Casdoor frontend from Ant Design to Tailwind CSS + shadcn/ui.

## Project Status

### Completed
- ✅ Tailwind CSS infrastructure setup
- ✅ shadcn/ui dependencies installed
- ✅ Core components created (Button, Input, Card, Table)
- ✅ Utility functions for className management
- ✅ Build configuration updated
- ✅ Reference implementation created

### Migration Statistics
- **Total Source Files**: 238
- **Files with Ant Design Imports**: 177
- **Estimated Migration Effort**: 4-6 weeks full-time

## Component Mapping

### Core UI Components

| Ant Design | shadcn/ui Equivalent | Status | Notes |
|-----------|---------------------|--------|-------|
| Button | Button | ✅ Created | Supports variants: default, primary, outline, ghost, destructive |
| Input | Input | ✅ Created | Direct replacement |
| Card | Card | ✅ Created | Includes CardHeader, CardTitle, CardContent, CardFooter |
| Table | Table | ✅ Created | Includes TableHeader, TableBody, TableRow, TableCell |
| Form | Form | ⏳ Pending | Use with React Hook Form |
| Select | Select | ⏳ Pending | Requires Radix UI primitive |
| Modal | Dialog | ⏳ Pending | Requires Radix UI primitive |
| Drawer | Sheet | ⏳ Pending | Requires Radix UI primitive |
| Dropdown | DropdownMenu | ⏳ Pending | Requires Radix UI primitive |
| Tooltip | Tooltip | ⏳ Pending | Requires Radix UI primitive |
| Alert | Alert | ⏳ Pending | Simple component to create |
| Badge | Badge | ⏳ Pending | Simple component to create |
| Tabs | Tabs | ⏳ Pending | Requires Radix UI primitive |
| Menu | NavigationMenu | ⏳ Pending | Requires Radix UI primitive |
| Checkbox | Checkbox | ⏳ Pending | Requires Radix UI primitive |
| Radio | RadioGroup | ⏳ Pending | Requires Radix UI primitive |
| Switch | Switch | ⏳ Pending | Requires Radix UI primitive |
| Slider | Slider | ⏳ Pending | Requires Radix UI primitive |
| DatePicker | Calendar | ⏳ Pending | Requires date-fns or similar |
| Pagination | Pagination | ⏳ Pending | Custom implementation needed |
| Spin | Spinner | ⏳ Pending | Simple component to create |

## Migration Strategy

### Phase 1: Infrastructure (✅ Completed)
1. Install Tailwind CSS and dependencies
2. Configure PostCSS and Tailwind
3. Install shadcn/ui dependencies
4. Create utility functions
5. Set up component directory structure

### Phase 2: Core Components (In Progress)
1. Create essential shadcn/ui components:
   - ✅ Button, Input, Card, Table
   - ⏳ Form, Select, Dialog (Modal)
   - ⏳ Alert, Badge, Tooltip
   - ⏳ Checkbox, Radio, Switch
   - ⏳ Dropdown, Tabs, Sheet (Drawer)

### Phase 3: Page Migration (Recommended Approach)
Migrate pages in this order to minimize risk:

1. **Authentication Pages** (High Impact, Moderate Complexity)
   - LoginPage.js
   - SignupPage.js
   - ForgetPage.js
   - MfaSetupPage.js

2. **Settings Pages** (Lower Impact, Lower Complexity)
   - Setting.js
   - SystemInfo.js

3. **List Pages** (High Complexity - Tables)
   - ApplicationListPage.js
   - UserListPage.js
   - OrganizationListPage.js
   - (Continue with other list pages)

4. **Edit Pages** (Highest Complexity - Forms)
   - ApplicationEditPage.js
   - UserEditPage.js
   - OrganizationEditPage.js
   - (Continue with other edit pages)

5. **Specialized Components**
   - Table components (table/ directory)
   - Modal components (common/modal/ directory)
   - Common utilities (common/ directory)

### Phase 4: Cleanup
1. Remove Ant Design imports
2. Remove Ant Design dependencies
3. Remove craco-less and LESS configuration
4. Clean up unused CSS
5. Update documentation

## Technical Considerations

### Dual System Approach
During migration, both Ant Design and Tailwind will coexist:
- Tailwind's `preflight` is disabled to prevent conflicts
- Gradually replace components page by page
- Test thoroughly after each migration

### Styling Approach
```javascript
// Use the cn() utility for conditional classes
import {cn} from "@/lib/utils";

<Button className={cn(
  "my-custom-class",
  isActive && "bg-blue-500",
  isDisabled && "opacity-50"
)}>
  Click Me
</Button>
```

### Form Handling
Recommend using React Hook Form instead of Ant Design Form:

```javascript
import {useForm} from "react-hook-form";
import {Input} from "@/components/ui/input";
import {Button} from "@/components/ui/button";

const MyForm = () => {
  const {register, handleSubmit} = useForm();
  
  const onSubmit = (data) => {
    console.log(data);
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Input {...register("email")} placeholder="Email" />
      <Button type="submit">Submit</Button>
    </form>
  );
};
```

## Example Migration

### Before (Ant Design)
```javascript
import {Button, Card, Table, Input} from "antd";

const MyComponent = () => (
  <Card title="Applications">
    <Input placeholder="Search" />
    <Button type="primary">Add New</Button>
    <Table dataSource={data} columns={columns} />
  </Card>
);
```

### After (Tailwind + shadcn/ui)
```javascript
import {Button} from "@/components/ui/button";
import {Card, CardHeader, CardTitle, CardContent} from "@/components/ui/card";
import {Table, TableHeader, TableBody, TableRow, TableCell} from "@/components/ui/table";
import {Input} from "@/components/ui/input";

const MyComponent = () => (
  <Card>
    <CardHeader>
      <CardTitle>Applications</CardTitle>
    </CardHeader>
    <CardContent>
      <div className="space-y-4">
        <Input placeholder="Search" />
        <Button variant="primary">Add New</Button>
        <Table>
          <TableHeader>{/* ... */}</TableHeader>
          <TableBody>{/* ... */}</TableBody>
        </Table>
      </div>
    </CardContent>
  </Card>
);
```

## Testing Strategy

### For Each Migrated Component
1. Visual regression testing
2. Functional testing (interactions work as expected)
3. Responsive design testing
4. Accessibility testing
5. Cross-browser testing

### Test Checklist
- [ ] Component renders correctly
- [ ] All interactive elements work
- [ ] Styling matches design system
- [ ] Responsive on mobile/tablet/desktop
- [ ] Keyboard navigation works
- [ ] Screen reader compatible
- [ ] No console errors or warnings

## Performance Considerations

### Bundle Size
- Tailwind CSS is more lightweight than Ant Design
- Tree-shaking ensures unused styles are removed
- Expected reduction in bundle size: 30-40%

### Runtime Performance
- No runtime style injection (unlike Ant Design)
- Smaller JavaScript bundle
- Faster initial page load

## Migration Checklist Template

For each page/component:

```markdown
## [Component/Page Name]

- [ ] Identify all Ant Design components used
- [ ] Create/verify shadcn/ui equivalents exist
- [ ] Migrate component code
- [ ] Update imports
- [ ] Test functionality
- [ ] Test responsive design
- [ ] Test accessibility
- [ ] Code review
- [ ] Update documentation
```

## Resources

- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [shadcn/ui Components](https://ui.shadcn.com)
- [Radix UI Primitives](https://www.radix-ui.com)
- [React Hook Form](https://react-hook-form.com)

## Getting Help

For questions or issues during migration:
1. Check this guide first
2. Review the reference implementation in `TailwindMigrationExample.js`
3. Consult shadcn/ui documentation
4. Ask in the Casdoor Discord

## Timeline Estimate

- **Phase 1** (Infrastructure): ✅ 1 day (Completed)
- **Phase 2** (Core Components): 3-5 days
- **Phase 3** (Page Migration): 3-4 weeks
- **Phase 4** (Cleanup): 2-3 days
- **Total**: 4-6 weeks for complete migration

## Notes

- This is a large-scale refactoring that changes the entire UI layer
- Take incremental approach to minimize risk
- Maintain feature parity throughout migration
- Consider creating a feature flag for gradual rollout
- Document any deviations from the original design
