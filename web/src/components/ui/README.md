# shadcn/ui Components

This directory contains the shadcn/ui components used in the Casdoor frontend migration from Ant Design to Tailwind CSS.

## Available Components

### ✅ Implemented

| Component | File | Description | Ant Design Equivalent |
|-----------|------|-------------|----------------------|
| Button | `button.js` | Button with multiple variants | `antd Button` |
| Input | `input.js` | Text input field | `antd Input` |
| Card | `card.js` | Card container with header, content, footer | `antd Card` |
| Table | `table.js` | Data table with header, body, rows, cells | `antd Table` |
| Dialog | `dialog.js` | Modal dialog with overlay | `antd Modal` |
| Alert | `alert.js` | Alert notifications with variants | `antd Alert` |
| Badge | `badge.js` | Badge/tag component with variants | `antd Badge` |

### ⏳ Recommended Next Components

| Component | Priority | Ant Design Equivalent | Notes |
|-----------|----------|----------------------|-------|
| Form | High | `antd Form` | Use with React Hook Form |
| Select | High | `antd Select` | Radix UI based |
| Checkbox | High | `antd Checkbox` | Radix UI based |
| Radio | High | `antd Radio` | Radix UI based |
| Switch | Medium | `antd Switch` | Radix UI based |
| Tabs | Medium | `antd Tabs` | Radix UI based |
| Dropdown | Medium | `antd Dropdown` | Radix UI based |
| Tooltip | Medium | `antd Tooltip` | Radix UI based |
| Sheet | Medium | `antd Drawer` | Radix UI based |
| Pagination | Low | `antd Pagination` | Custom implementation |
| DatePicker | Low | `antd DatePicker` | Requires date library |

## Usage

### Import Components

```javascript
import {Button} from "./components/ui/button";
import {Card, CardHeader, CardTitle, CardContent} from "./components/ui/card";
import {Dialog, DialogTrigger, DialogContent} from "./components/ui/dialog";
```

### Button Example

```javascript
<Button variant="primary">Click Me</Button>
<Button variant="outline">Secondary</Button>
<Button variant="destructive">Delete</Button>
<Button variant="ghost">Cancel</Button>
```

### Card Example

```javascript
<Card>
  <CardHeader>
    <CardTitle>Title</CardTitle>
    <CardDescription>Description text</CardDescription>
  </CardHeader>
  <CardContent>
    Main content goes here
  </CardContent>
  <CardFooter>
    Footer content
  </CardFooter>
</Card>
```

### Dialog Example

```javascript
const [open, setOpen] = useState(false);

<Dialog open={open} onOpenChange={setOpen}>
  <DialogTrigger asChild>
    <Button>Open Dialog</Button>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Dialog Title</DialogTitle>
      <DialogDescription>Dialog description</DialogDescription>
    </DialogHeader>
    <div>Dialog content</div>
    <DialogFooter>
      <Button onClick={() => setOpen(false)}>Close</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

### Table Example

```javascript
<Table>
  <TableHeader>
    <TableRow>
      <TableHead>Column 1</TableHead>
      <TableHead>Column 2</TableHead>
    </TableRow>
  </TableHeader>
  <TableBody>
    {data.map((item) => (
      <TableRow key={item.id}>
        <TableCell>{item.name}</TableCell>
        <TableCell>{item.value}</TableCell>
      </TableRow>
    ))}
  </TableBody>
</Table>
```

### Alert Example

```javascript
<Alert variant="success">
  <CheckCircle className="h-4 w-4" />
  <AlertTitle>Success</AlertTitle>
  <AlertDescription>Your action was successful.</AlertDescription>
</Alert>
```

### Badge Example

```javascript
<Badge variant="success">Active</Badge>
<Badge variant="warning">Pending</Badge>
<Badge variant="destructive">Error</Badge>
```

## Component Structure

All components follow the shadcn/ui pattern:

1. **Built on Radix UI** (for complex components): Provides accessibility, keyboard navigation, and focus management
2. **Styled with Tailwind**: Uses utility classes for styling
3. **Customizable**: Uses `class-variance-authority` for variant management
4. **Composable**: Components are exported as React components that can be composed together

## Styling

### Using the cn() Utility

```javascript
import {cn} from "../../lib/utils";

<Button className={cn(
  "my-custom-class",
  isActive && "bg-blue-500",
  isDisabled && "opacity-50"
)}>
  Click Me
</Button>
```

### Tailwind Classes

Components use Tailwind's utility classes:
- Spacing: `p-4`, `m-2`, `space-y-4`
- Colors: `bg-blue-500`, `text-white`
- Borders: `border`, `rounded-md`
- Flexbox: `flex`, `items-center`, `justify-between`

## Accessibility

All components are designed with accessibility in mind:
- Proper ARIA attributes
- Keyboard navigation support
- Focus management
- Screen reader compatibility

## Customization

### Extending Components

You can extend components by:

1. **Adding new variants:**
```javascript
const buttonVariants = cva(
  "base-classes",
  {
    variants: {
      variant: {
        // Add your custom variant
        custom: "bg-purple-500 text-white hover:bg-purple-600",
      },
    },
  }
);
```

2. **Creating wrapper components:**
```javascript
const PrimaryButton = ({children, ...props}) => (
  <Button variant="primary" {...props}>
    {children}
  </Button>
);
```

## Adding New Components

To add a new shadcn/ui component:

1. Check if it requires a Radix UI primitive
2. Install the Radix UI package if needed:
   ```bash
   yarn add @radix-ui/react-[component-name]
   ```
3. Create the component file in this directory
4. Follow the shadcn/ui pattern:
   - Use `forwardRef` for ref forwarding
   - Use `cn()` utility for className merging
   - Use `cva()` for variant management
   - Export all component parts

## Resources

- [shadcn/ui Documentation](https://ui.shadcn.com)
- [Radix UI Primitives](https://www.radix-ui.com)
- [Tailwind CSS](https://tailwindcss.com)
- [class-variance-authority](https://cva.style)

## Migration Guide

See `../../MIGRATION_GUIDE.md` for detailed migration instructions and component mapping.
