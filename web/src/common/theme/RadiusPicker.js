import {InputNumber, Slider, Space} from "antd";

export default function RadiusPicker({value, onChange}) {
  return (
    <Space size="large">
      <InputNumber
        value={value}
        onChange={onChange}
        style={{width: 120}}
        min={0}
        formatter={(val) => `${val}px`}
        parser={(str) => (str ? parseFloat(str) : str)}
      />

      <Slider
        tooltip={{open: false}}
        style={{width: 128}}
        min={0}
        value={value}
        max={20}
        onChange={onChange}
      />
    </Space>
  );
}
