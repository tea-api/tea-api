import React, { useContext, useEffect, useRef, useMemo, useState } from 'react';
import { API, copy, showError, showInfo, showSuccess } from '../helpers';
import { useTranslation } from 'react-i18next';

import {
  Banner,
  Input,
  Layout,
  Modal,
  Space,
  Table,
  Tag,
  Tooltip,
  Popover,
  ImagePreview,
  RadioGroup,
  Radio,
  Button,
} from '@douyinfe/semi-ui';
import {
  IconMore,
  IconVerify,
  IconUploadError,
  IconHelpCircle,
  IconEdit,
} from '@douyinfe/semi-icons';
import { UserContext } from '../context/User/index.js';
import { isAdmin } from '../helpers/utils.js';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';

const ModelPricing = () => {
  const { t } = useTranslation();
  const [filteredValue, setFilteredValue] = useState([]);
  const compositionRef = useRef({ isComposition: false });
  const [selectedRowKeys, setSelectedRowKeys] = useState([]);
  const [modalImageUrl, setModalImageUrl] = useState('');
  const [isModalOpenurl, setIsModalOpenurl] = useState(false);
  const [selectedGroup, setSelectedGroup] = useState('default');
  const [editVisible, setEditVisible] = useState(false);
  const [editModel, setEditModel] = useState(null);
  const [inputPrice, setInputPrice] = useState('');
  const [outputPrice, setOutputPrice] = useState('');
  const [priceUnit, setPriceUnit] = useState('1k');

  const rowSelection = useMemo(
    () => ({
      onChange: (selectedRowKeys, selectedRows) => {
        setSelectedRowKeys(selectedRowKeys);
      },
    }),
    [],
  );

  const handleChange = (value) => {
    if (compositionRef.current.isComposition) {
      return;
    }
    const newFilteredValue = value ? [value] : [];
    setFilteredValue(newFilteredValue);
  };
  const handleCompositionStart = () => {
    compositionRef.current.isComposition = true;
  };

  const handleCompositionEnd = (event) => {
    compositionRef.current.isComposition = false;
    const value = event.target.value;
    const newFilteredValue = value ? [value] : [];
    setFilteredValue(newFilteredValue);
  };

  const openEdit = (record) => {
    // 先设置单位为默认值
    setPriceUnit('1k');
    
    // 使用与后端一致的基准价格计算方式
    const basePerM = 2.0;
    const base = basePerM / 1000; // 转换为每千 tokens 的价格
    
    // 根据模型类型设置初始价格
    if (record.quota_type === 0) {
      // 按量计费模型
      // 显示原始倍率对应的价格
      setInputPrice((record.model_ratio * base).toFixed(3));
      setOutputPrice((record.model_ratio * record.completion_ratio * base).toFixed(3));
    } else {
      // 按次计费模型
      setInputPrice(record.model_price.toFixed(3));
      setOutputPrice('0');
    }
    
    setEditModel(record);
    setEditVisible(true);
  };

  const savePrice = async () => {
    if (!editModel) return;
    const res = await API.put('/api/pricing', {
      model_name: editModel.model_name,
      input_price: parseFloat(inputPrice),
      output_price: parseFloat(outputPrice),
      unit: priceUnit,
    });
    if (res.data.success) {
      showSuccess(t('保存成功'));
      setEditVisible(false);
      await refresh();
    } else {
      showError(res.data.message);
    }
  };

  function renderQuotaType(type) {
    // Ensure all cases are string literals by adding quotes.
    switch (type) {
      case 1:
        return (
          <Tag color='teal' size='large'>
            {t('按次计费')}
          </Tag>
        );
      case 0:
        return (
          <Tag color='violet' size='large'>
            {t('按量计费')}
          </Tag>
        );
      default:
        return t('未知');
    }
  }

  function renderAvailable(available) {
    return available ? (
      <Popover
        content={
          <div style={{ padding: 8 }}>{t('您的分组可以使用该模型')}</div>
        }
        position='top'
        key={available}
        style={{
          backgroundColor: 'rgba(var(--semi-blue-4),1)',
          borderColor: 'rgba(var(--semi-blue-4),1)',
          color: 'var(--semi-color-white)',
          borderWidth: 1,
          borderStyle: 'solid',
        }}
      >
        <IconVerify style={{ color: 'green' }} size='large' />
      </Popover>
    ) : null;
  }

  const columns = [
    {
      title: t('可用性'),
      dataIndex: 'available',
      render: (text, record, index) => {
        // if record.enable_groups contains selectedGroup, then available is true
        return renderAvailable(record.enable_groups.includes(selectedGroup));
      },
      sorter: (a, b) => {
        const aAvailable = a.enable_groups.includes(selectedGroup);
        const bAvailable = b.enable_groups.includes(selectedGroup);
        return Number(aAvailable) - Number(bAvailable);
      },
      defaultSortOrder: 'descend',
    },
    {
      title: t('模型名称'),
      dataIndex: 'model_name',
      render: (text, record, index) => {
        return (
          <>
            <Tag
              color='green'
              size='large'
              onClick={() => {
                copyText(text);
              }}
            >
              {text}
            </Tag>
          </>
        );
      },
      onFilter: (value, record) =>
        record.model_name.toLowerCase().includes(value.toLowerCase()),
      filteredValue,
    },
    {
      title: t('计费类型'),
      dataIndex: 'quota_type',
      render: (text, record, index) => {
        return renderQuotaType(parseInt(text));
      },
      sorter: (a, b) => a.quota_type - b.quota_type,
    },
    {
      title: t('可用分组'),
      dataIndex: 'enable_groups',
      render: (text, record, index) => {
        // enable_groups is a string array
        return (
          <Space>
            {text.map((group) => {
              if (usableGroup[group]) {
                if (group === selectedGroup) {
                  return (
                    <Tag color='blue' size='large' prefixIcon={<IconVerify />}>
                      {group}
                    </Tag>
                  );
                } else {
                  return (
                    <Tag
                      color='blue'
                      size='large'
                      onClick={() => {
                        setSelectedGroup(group);
                        showInfo(
                          t('当前查看的分组为：{{group}}，倍率为：{{ratio}}', {
                            group: group,
                            ratio: groupRatio[group],
                          }),
                        );
                      }}
                    >
                      {group}
                    </Tag>
                  );
                }
              }
            })}
          </Space>
        );
      },
    },
    {
      title: () => (
        <span style={{ display: 'flex', alignItems: 'center' }}>
          {t('倍率')}
          <Popover
            content={
              <div style={{ padding: 8 }}>
                {t('倍率是为了方便换算不同价格的模型')}
                <br />
                {t('点击查看倍率说明')}
              </div>
            }
            position='top'
            style={{
              backgroundColor: 'rgba(var(--semi-blue-4),1)',
              borderColor: 'rgba(var(--semi-blue-4),1)',
              color: 'var(--semi-color-white)',
              borderWidth: 1,
              borderStyle: 'solid',
            }}
          >
            <IconHelpCircle
              onClick={() => {
                setModalImageUrl('/ratio.png');
                setIsModalOpenurl(true);
              }}
            />
          </Popover>
        </span>
      ),
      dataIndex: 'model_ratio',
      render: (text, record, index) => {
        let content = text;
        let completionRatio = parseFloat(record.completion_ratio.toFixed(3));
        content = (
          <>
            <Text>
              {t('模型倍率')}：{record.quota_type === 0 ? text : t('无')}
            </Text>
            <br />
            <Text>
              {t('补全倍率')}：
              {record.quota_type === 0 ? completionRatio : t('无')}
            </Text>
            <br />
            <Text>
              {t('分组倍率')}：{groupRatio[selectedGroup]}
            </Text>
          </>
        );
        return <div>{content}</div>;
      },
    },
    {
      title: t('模型价格'),
      dataIndex: 'model_price',
      render: (text, record, index) => {
        let content = text;
        if (record.quota_type === 0) {
          // 按量计费模型
          // 使用与后端一致的基准价格计算方式 - 1倍率=0.002刀
          const basePerM = 2.0;
          const base = basePerM / 1000; // 转换为每千 tokens 的价格
          
          let inputRatioPrice = record.model_ratio * base * groupRatio[selectedGroup];
          let completionRatioPrice = record.model_ratio * record.completion_ratio * base * groupRatio[selectedGroup];
          
          // 显示为每1M tokens的价格
          let inputRatioPricePerM = inputRatioPrice * 1000;
          let completionRatioPricePerM = completionRatioPrice * 1000;
          
          content = (
            <>
              <Text>
                {t('提示')} ${inputRatioPricePerM.toFixed(3)} / 1M tokens
              </Text>
              <br />
              <Text>
                {t('补全')} ${completionRatioPricePerM.toFixed(3)} / 1M tokens
              </Text>
            </>
          );
        } else {
          // 按次计费模型
          let price = parseFloat(text) * groupRatio[selectedGroup];
          content = (
            <>
              {t('模型价格')}：${price.toFixed(3)}
            </>
          );
        }
        return <div>{content}</div>;
      },
    },
    {
      title: t('编辑价格'),
      dataIndex: 'action',
      render: (text, record) => (
        isAdmin() && (
          <Button icon={<IconEdit />} onClick={() => openEdit(record)} />
        )
      ),
    },
  ];

  const [models, setModels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [userState, userDispatch] = useContext(UserContext);
  const [groupRatio, setGroupRatio] = useState({});
  const [usableGroup, setUsableGroup] = useState({});

  const setModelsFormat = (models, groupRatio) => {
    for (let i = 0; i < models.length; i++) {
      models[i].key = models[i].model_name;
      models[i].group_ratio = groupRatio[models[i].model_name];
    }
    // sort by quota_type
    models.sort((a, b) => {
      return a.quota_type - b.quota_type;
    });

    // sort by model_name, start with gpt is max, other use localeCompare
    models.sort((a, b) => {
      if (a.model_name.startsWith('gpt') && !b.model_name.startsWith('gpt')) {
        return -1;
      } else if (
        !a.model_name.startsWith('gpt') &&
        b.model_name.startsWith('gpt')
      ) {
        return 1;
      } else {
        return a.model_name.localeCompare(b.model_name);
      }
    });

    setModels(models);
  };

  const loadPricing = async () => {
    setLoading(true);

    let url = '';
    url = `/api/pricing`;
    const res = await API.get(url);
    const { success, message, data, group_ratio, usable_group } = res.data;
    if (success) {
      setGroupRatio(group_ratio);
      setUsableGroup(usable_group);
      setSelectedGroup(userState.user ? userState.user.group : 'default');
      setModelsFormat(data, group_ratio);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const refresh = async () => {
    await loadPricing();
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess('已复制：' + text);
    } else {
      // setSearchKeyword(text);
      Modal.error({ title: '无法复制到剪贴板，请手动复制', content: text });
    }
  };

  useEffect(() => {
    refresh().then();
  }, []);

  // 处理输入价格变化
  const handleInputPriceChange = (value) => {
    setInputPrice(value);
    // 计算补全倍率，但不显示提示
    const inputVal = parseFloat(value);
    const outputVal = parseFloat(outputPrice);
    
    // 保留计算逻辑，但删除showInfo
  };
  
  // 处理输出价格变化
  const handleOutputPriceChange = (value) => {
    setOutputPrice(value);
    // 计算补全倍率，但不显示提示
    const inputVal = parseFloat(inputPrice);
    const outputVal = parseFloat(value);
    
    // 保留计算逻辑，但删除showInfo
  };

  // 处理单位切换
  const handleUnitChange = (value) => {
    if (!editModel) return;
    
    const oldUnit = priceUnit;
    const newUnit = value;
    
    // 如果单位不同，需要转换价格
    if (oldUnit !== newUnit) {
      const currentInputPrice = parseFloat(inputPrice);
      const currentOutputPrice = parseFloat(outputPrice);
      
      let newInputPrice, newOutputPrice;
      
      if (oldUnit === '1k' && newUnit === '1m') {
        // 从1k转换到1M，价格乘以1000
        newInputPrice = (currentInputPrice * 1000).toFixed(3);
        newOutputPrice = (currentOutputPrice * 1000).toFixed(3);
      } else if (oldUnit === '1m' && newUnit === '1k') {
        // 从1M转换到1k，价格除以1000
        newInputPrice = (currentInputPrice / 1000).toFixed(3);
        newOutputPrice = (currentOutputPrice / 1000).toFixed(3);
      }
      
      setInputPrice(newInputPrice);
      setOutputPrice(newOutputPrice);
      
      // 保留计算逻辑，但删除showInfo
    }
    
    setPriceUnit(value);
  };

  return (
    <>
      <Layout>
        {userState.user ? (
          <Banner
            type='success'
            fullMode={false}
            closeIcon='null'
            description={t('您的默认分组为：{{group}}，分组倍率为：{{ratio}}', {
              group: userState.user.group,
              ratio: groupRatio[userState.user.group],
            })}
          />
        ) : (
          <Banner
            type='warning'
            fullMode={false}
            closeIcon='null'
            description={t('您还未登陆，显示的价格为默认分组倍率: {{ratio}}', {
              ratio: groupRatio['default'],
            })}
          />
        )}
        <br />
        <Banner
          type='info'
          fullMode={false}
          description={
            <div>
              {t(
                '按量计费费用 = 分组倍率 × 模型倍率 × （提示token数 + 补全token数 × 补全倍率）/ 500000 （单位：美元）',
              )}
            </div>
          }
          closeIcon='null'
        />
        <br />
        <Space style={{ marginBottom: 16 }}>
          <Input
            placeholder={t('模糊搜索模型名称')}
            style={{ width: 200 }}
            onCompositionStart={handleCompositionStart}
            onCompositionEnd={handleCompositionEnd}
            onChange={handleChange}
            showClear
          />
          <Button
            theme='light'
            type='tertiary'
            style={{ width: 150 }}
            onClick={() => {
              copyText(selectedRowKeys);
            }}
            disabled={selectedRowKeys == ''}
          >
            {t('复制选中模型')}
          </Button>
        </Space>
        <Table
          style={{ marginTop: 5 }}
          columns={columns}
          dataSource={models}
          loading={loading}
          pagination={{
            formatPageText: (page) =>
              t('第 {{start}} - {{end}} 条，共 {{total}} 条', {
                start: page.currentStart,
                end: page.currentEnd,
                total: models.length,
              }),
            pageSize: models.length,
            showSizeChanger: false,
          }}
          rowSelection={rowSelection}
        />
        <Modal
          visible={editVisible}
          title={t('编辑价格')}
          onOk={savePrice}
          onCancel={() => setEditVisible(false)}
        >
          <RadioGroup
            type='button'
            value={priceUnit}
            onChange={(e) => handleUnitChange(e.target.value)}
            style={{ marginBottom: 12 }}
          >
            <Radio value='1k'>/1k tokens</Radio>
            <Radio value='1m'>/1M tokens</Radio>
          </RadioGroup>
          <Input
            value={inputPrice}
            onChange={handleInputPriceChange}
            suffix={`$/1${priceUnit === '1k' ? 'k' : 'M'} tokens`}
            label={t('实际输入价格')}
          />
          <Input
            value={outputPrice}
            onChange={handleOutputPriceChange}
            suffix={`$/1${priceUnit === '1k' ? 'k' : 'M'} tokens`}
            label={t('实际输出价格')}
            style={{ marginTop: 12 }}
          />
          {(() => {
            const inputVal = parseFloat(inputPrice);
            const outputVal = parseFloat(outputPrice);
            
            if (!isNaN(inputVal) && !isNaN(outputVal) && inputVal > 0) {
              const ratio = outputVal / inputVal;
              if (inputVal === outputVal) {
                return (
                  <Banner
                    type='info'
                    description={t('当前补全倍率为: 1')}
                    style={{ marginTop: 12 }}
                  />
                );
              } else {
                return (
                  <Banner
                    type='info'
                    description={t('当前补全倍率为: {{ratio}}', {
                      ratio: ratio.toFixed(3)
                    })}
                    style={{ marginTop: 12 }}
                  />
                );
              }
            }
            return null;
          })()}
        </Modal>
        <ImagePreview
          src={modalImageUrl}
          visible={isModalOpenurl}
          onVisibleChange={(visible) => setIsModalOpenurl(visible)}
        />
      </Layout>
    </>
  );
};

export default ModelPricing;
