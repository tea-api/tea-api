import React, { useEffect, useState } from 'react';
import {
  API,
  copy,
  showError,
  showSuccess,
  timestamp2string,
  downloadTextAsFile,
} from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import { renderQuota } from '../helpers/render';
import {
  Button,
  Divider,
  Form,
  Modal,
  Popconfirm,
  Popover,
  Table,
  Tag,
  Input,
} from '@douyinfe/semi-ui';
import EditRedemption from '../pages/Redemption/EditRedemption';
import { useTranslation } from 'react-i18next';

function renderTimestamp(timestamp) {
  return <>{timestamp2string(timestamp)}</>;
}

const RedemptionsTable = () => {
  const { t } = useTranslation();

  const renderStatus = (status) => {
    switch (status) {
      case 1:
        return (
          <Tag color='green' size='large'>
            {t('未使用')}
          </Tag>
        );
      case 2:
        return (
          <Tag color='red' size='large'>
            {t('已禁用')}
          </Tag>
        );
      case 3:
        return (
          <Tag color='grey' size='large'>
            {t('已使用')}
          </Tag>
        );
      default:
        return (
          <Tag color='black' size='large'>
            {t('未知状态')}
          </Tag>
        );
    }
  };

  const columns = [
    {
      title: t('ID'),
      dataIndex: 'id',
    },
    {
      title: t('名称'),
      dataIndex: 'name',
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      key: 'status',
      render: (text, record, index) => {
        return <div>{renderStatus(text)}</div>;
      },
    },
    {
      title: t('额度'),
      dataIndex: 'quota',
      render: (text, record, index) => {
        return <div>{renderQuota(parseInt(text))}</div>;
      },
    },
    {
      title: t('最多使用次数'),
      dataIndex: 'max_times',
      render: (text, record) => {
        return (
          <Popover content={t('每个兑换码可被使用的最大次数')} position='top'>
            <div>{text}</div>
          </Popover>
        );
      },
    },
    {
      title: t('每用户最多使用次数'),
      dataIndex: 'max_user_times',
      render: (text, record) => {
        return (
          <Popover content={t('每个用户最多可使用该兑换码的次数')} position='top'>
            <div>{text}</div>
          </Popover>
        );
      },
    },
    {
      title: t('已使用次数'),
      dataIndex: 'used_times',
    },
    {
      title: t('过期时间'),
      dataIndex: 'expired_time',
      render: (text, record) => {
        return (
          <div>
            {record.expired_time === -1 ? t('永不过期') : renderTimestamp(text)}
          </div>
        );
      },
    },
    {
      title: t('创建时间'),
      dataIndex: 'created_time',
      render: (text, record, index) => {
        return <div>{renderTimestamp(text)}</div>;
      },
    },
    {
      title: t('兑换人ID'),
      dataIndex: 'used_user_id',
      render: (text, record, index) => {
        return <div>{text === 0 ? t('无') : text}</div>;
      },
    },
    {
      title: '',
      dataIndex: 'operate',
      render: (text, record, index) => (
        <div>
          <Popover content={record.key} style={{ padding: 20 }} position='top'>
            <Button theme='light' type='tertiary' style={{ marginRight: 1 }}>
              {t('查看')}
            </Button>
          </Popover>
          <Button
            theme='light'
            type='secondary'
            style={{ marginRight: 1 }}
            onClick={async (text) => {
              await copyText(record.key);
            }}
          >
            {t('复制')}
          </Button>
          <Popconfirm
            title={t('确定是否要删除此兑换码？')}
            content={t('此修改将不可逆')}
            okType={'danger'}
            position={'left'}
            onConfirm={() => {
              manageRedemption(record.id, 'delete', record).then(() => {
                removeRecord(record.key);
              });
            }}
          >
            <Button theme='light' type='danger' style={{ marginRight: 1 }}>
              {t('删除')}
            </Button>
          </Popconfirm>
          {record.status === 1 ? (
            <Button
              theme='light'
              type='warning'
              style={{ marginRight: 1 }}
              onClick={async () => {
                manageRedemption(record.id, 'disable', record);
              }}
            >
              {t('禁用')}
            </Button>
          ) : (
            <Button
              theme='light'
              type='secondary'
              style={{ marginRight: 1 }}
              onClick={async () => {
                manageRedemption(record.id, 'enable', record);
              }}
              disabled={record.status === 3}
            >
              {t('启用')}
            </Button>
          )}
          <Button
            theme='light'
            type='tertiary'
            style={{ marginRight: 1 }}
            onClick={() => {
              setEditingRedemption(record);
              setShowEdit(true);
            }}
            disabled={record.status !== 1}
          >
            {t('编辑')}
          </Button>
        </div>
      ),
    },
  ];

  const [redemptions, setRedemptions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [tokenCount, setTokenCount] = useState(ITEMS_PER_PAGE);
  const [selectedKeys, setSelectedKeys] = useState([]);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [editingRedemption, setEditingRedemption] = useState({
    id: undefined,
  });
  const [showEdit, setShowEdit] = useState(false);

  const closeEdit = () => {
    setShowEdit(false);
  };

  const setRedemptionFormat = (redeptions) => {
    setRedemptions(redeptions);
  };

  const loadRedemptions = async (startIdx, pageSize) => {
    const res = await API.get(
      `/api/redemption/?p=${startIdx}&page_size=${pageSize}`,
    );
    const { success, message, data } = res.data;
    if (success) {
      const newPageData = data.items;
      setActivePage(data.page);
      setTokenCount(data.total);
      setRedemptionFormat(newPageData);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const removeRecord = (key) => {
    let newDataSource = [...redemptions];
    if (key != null) {
      let idx = newDataSource.findIndex((data) => data.key === key);

      if (idx > -1) {
        newDataSource.splice(idx, 1);
        setRedemptions(newDataSource);
      }
    }
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess(t('已复制到剪贴板！'));
    } else {
      // setSearchKeyword(text);
      Modal.error({ title: t('无法复制到剪贴板，请手动复制'), content: text });
    }
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(redemptions.length / pageSize) + 1) {
        await loadRedemptions(activePage - 1, pageSize);
      }
      setActivePage(activePage);
    })();
  };

  useEffect(() => {
    loadRedemptions(0, pageSize)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []);

  const refresh = async () => {
    setLoading(true);
    await loadRedemptions(1, pageSize);
  };

  const manageRedemption = async (id, action, record) => {
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/redemption/${id}/`);
        break;
      case 'enable':
        data.status = 1;
        res = await API.put('/api/redemption/?status_only=true', data);
        break;
      case 'disable':
        data.status = 2;
        res = await API.put('/api/redemption/?status_only=true', data);
        break;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('操作成功完成！'));
      let redemption = res.data.data;
      let newRedemptions = [...redemptions];
      // let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      if (action === 'delete') {
      } else {
        record.status = redemption.status;
      }
      setRedemptions(newRedemptions);
    } else {
      showError(message);
    }
  };

  const searchRedemptions = async (keyword, page, pageSize) => {
    if (searchKeyword === '') {
      await loadRedemptions(page, pageSize);
      return;
    }
    setSearching(true);
    const res = await API.get(
      `/api/redemption/search?keyword=${keyword}&p=${page}&page_size=${pageSize}`,
    );
    const { success, message, data } = res.data;
    if (success) {
      const newPageData = data.items;
      setActivePage(data.page);
      setTokenCount(data.total);
      setRedemptionFormat(newPageData);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (value) => {
    setSearchKeyword(value.trim());
  };

  const sortRedemption = (key) => {
    if (redemptions.length === 0) return;
    setLoading(true);
    let sortedRedemptions = [...redemptions];
    sortedRedemptions.sort((a, b) => {
      return ('' + a[key]).localeCompare(b[key]);
    });
    if (sortedRedemptions[0].id === redemptions[0].id) {
      sortedRedemptions.reverse();
    }
    setRedemptions(sortedRedemptions);
    setLoading(false);
  };

  const handlePageChange = (page) => {
    setActivePage(page);
    if (searchKeyword === '') {
      loadRedemptions(page, pageSize).then();
    } else {
      searchRedemptions(searchKeyword, page, pageSize).then();
    }
  };

  let pageData = redemptions;
  const rowSelection = {
    onSelect: (record, selected) => {},
    onSelectAll: (selected, selectedRows) => {},
    onChange: (selectedRowKeys, selectedRows) => {
      setSelectedKeys(selectedRows);
    },
  };

  const handleRow = (record, index) => {
    if (record.status !== 1) {
      return {
        style: {
          background: 'var(--semi-color-disabled-border)',
        },
      };
    } else {
      return {};
    }
  };

  useEffect(() => {
    refresh();
  }, []);

  const batchGenerateRedemptions = () => {
    let formApi;
    Modal.confirm({
      title: t('批量生成兑换码'),
      content: (
        <div>
          <Form getFormApi={(api) => { formApi = api; }}>
            <Form.Input 
              field="prefix" 
              label={t('兑换码前缀')} 
              placeholder={t('例如：GIFT2024')} 
              initValue="GIFT"
            />
            <Form.InputNumber 
              field="count" 
              label={t('生成数量')} 
              placeholder={t('1-100')} 
              initValue={10}
              min={1} 
              max={100}
            />
            <Form.InputNumber 
              field="quota" 
              label={t('额度')} 
              placeholder={t('兑换码额度')} 
              initValue={100000}
            />
            <Form.InputNumber 
              field="maxTimes" 
              label={t('最多使用次数')} 
              placeholder={t('每个兑换码可使用次数')} 
              initValue={1}
              min={1}
            />
            <Form.InputNumber 
              field="maxUserTimes" 
              label={t('每用户最多使用次数')} 
              placeholder={t('每个用户最多可使用该兑换码的次数')} 
              initValue={1}
              min={1}
            />
          </Form>
        </div>
      ),
      onOk: async () => {
        const values = formApi.getValues();
        if (!values.prefix) {
          showError(t('请输入兑换码前缀'));
          return;
        }
        const res = await API.post('/api/redemption/', {
          name: values.prefix,
          key: values.prefix + '*',
          count: values.count || 10,
          quota: values.quota || 100000,
          max_times: values.maxTimes || 1,
          max_user_times: values.maxUserTimes || 1,
          expired_time: -1
        });
        const { success, message, data } = res.data;
        if (success) {
          showSuccess(t('兑换码批量生成成功！'));
          let text = '';
          for (let i = 0; i < data.length; i++) {
            text += data[i] + '\n';
          }
          Modal.confirm({
            title: t('兑换码创建成功'),
            content: (
              <div>
                <p>{t('兑换码创建成功，是否下载兑换码？')}</p>
                <p>{t('兑换码将以文本文件的形式下载，文件名为兑换码的前缀。')}</p>
              </div>
            ),
            onOk: () => {
              downloadTextAsFile(text, `${values.prefix}.txt`);
            },
          });
          refresh();
        } else {
          showError(message);
        }
      },
    });
  };

  return (
    <>
      <EditRedemption
        refresh={refresh}
        editingRedemption={editingRedemption}
        visiable={showEdit}
        handleClose={closeEdit}
      ></EditRedemption>
      <Form
        onSubmit={() => {
          searchRedemptions(searchKeyword, activePage, pageSize).then();
        }}
      >
        <Form.Input
          label={t('搜索关键字')}
          field='keyword'
          icon='search'
          iconPosition='left'
          placeholder={t('关键字(id或者名称)')}
          value={searchKeyword}
          loading={searching}
          onChange={handleKeywordChange}
        />
      </Form>
      <Divider style={{ margin: '5px 0 15px 0' }} />
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '16px' }}>
        <div style={{ display: 'flex', gap: '10px' }}>
          <Button theme='light' type='primary' style={{ marginRight: 8 }} onClick={() => {
            setEditingRedemption({});
            setShowEdit(true);
          }}>
            {t('添加')}
          </Button>
          <Button theme='light' type='secondary' onClick={batchGenerateRedemptions}>
            {t('批量生成')}
          </Button>
        </div>
        <div>
          <Input
            style={{ width: 300 }}
            showClear
            prefix={<i className="fa fa-search" aria-hidden="true"></i>}
            placeholder={t('搜索兑换码')}
            onChange={(value) => handleKeywordChange(value)}
            value={searchKeyword}
          />
        </div>
      </div>
      <Table
        style={{ marginTop: 20 }}
        columns={columns}
        dataSource={pageData}
        pagination={{
          currentPage: activePage,
          pageSize: pageSize,
          total: tokenCount,
          showSizeChanger: true,
          pageSizeOpts: [10, 20, 50, 100],
          formatPageText: (page) =>
            t('第 {{start}} - {{end}} 条，共 {{total}} 条', {
              start: page.currentStart,
              end: page.currentEnd,
              total: tokenCount,
            }),
          onPageSizeChange: (size) => {
            setPageSize(size);
            setActivePage(1);
            if (searchKeyword === '') {
              loadRedemptions(1, size).then();
            } else {
              searchRedemptions(searchKeyword, 1, size).then();
            }
          },
          onPageChange: handlePageChange,
        }}
        loading={loading}
        rowSelection={rowSelection}
        onRow={handleRow}
      ></Table>
    </>
  );
};

export default RedemptionsTable;
