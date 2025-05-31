import React from 'react';
import { Input, Button, Space } from '@douyinfe/semi-ui';
import { IconDelete } from '@douyinfe/semi-icons';

const KeyValueList = ({
  pairs,
  onChange,
  addText,
  keyPlaceholder,
  valuePlaceholder,
}) => {
  const updatePairKey = (index, key) => {
    const newPairs = [...pairs];
    newPairs[index].key = key;
    onChange(newPairs);
  };
  const updatePairValue = (index, value) => {
    const newPairs = [...pairs];
    newPairs[index].value = value;
    onChange(newPairs);
  };
  const addPair = () => {
    onChange([...pairs, { key: '', value: '' }]);
  };
  const removePair = (index) => {
    const newPairs = pairs.filter((_, i) => i !== index);
    onChange(newPairs);
  };
  return (
    <>
      {pairs.map((pair, index) => (
        <Space key={index} style={{ marginTop: 8 }}>
          <Input
            placeholder={keyPlaceholder}
            value={pair.key}
            onChange={(val) => updatePairKey(index, val)}
            style={{ width: 150 }}
          />
          <Input
            placeholder={valuePlaceholder}
            value={pair.value}
            onChange={(val) => updatePairValue(index, val)}
            style={{ width: 150 }}
          />
          <Button
            icon={<IconDelete />}
            theme='borderless'
            type='danger'
            onClick={() => removePair(index)}
          />
        </Space>
      ))}
      <div style={{ marginTop: 8 }}>
        <Button onClick={addPair} size='small'>
          {addText}
        </Button>
      </div>
    </>
  );
};

export default KeyValueList;
